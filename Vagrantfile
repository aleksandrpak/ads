# -*- mode: ruby -*-
# vi: set ft=ruby :

require 'yaml'

VAGRANTFILE_API_VERSION = "2"

settings = YAML.load_file("#{File.dirname(__FILE__)}/vagrant.yaml")

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  $vm_name = settings["vmname"] ||= "vagrant"
  $host_name = settings["hostname"] ||= "vagrant.dev"
  $int_ip = settings["ip"] ||= "192.168.56.206"
  $ext_ip = Socket::getaddrinfo($host_name, 'http', nil, Socket::SOCK_STREAM)[0][3]

  config.vm.provider :virtualbox do |vb|
    vb.name = $vm_name

    vb.customize ["setextradata", :id, "VBoxInternal2/SharedFoldersEnableSymlinksCreate/v-root", "1"]

    vb.customize ["modifyvm", :id, "--memory", settings["memory"] ||= "256"]
    vb.customize ["modifyvm", :id, "--cpus", settings["cpus"] ||= "1"]
    vb.customize ["modifyvm", :id, "--ioapic", "on"]
    vb.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
    vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
    vb.customize ["modifyvm", :id, "--ostype", "Ubuntu_64"]
  end

  config.vm.box = "ubuntu/trusty64"
  config.vm.hostname = $host_name
  config.vm.network :private_network, ip: $int_ip

  if not $ext_ip.eql?($int_ip)
    config.vm.network :public_network, ip: $ext_ip, bridge: [
      "en0: Wi-Fi (AirPort)"
    ]
  else
    config.vm.network :public_network, auto_config: false, bridge: [
      "en0: Wi-Fi (AirPort)",
      "en1: Wi-Fi (AirPort)",
      "wlan0",
      "eth0"
    ]
    config.vm.provision "shell",
      run: "always",
      inline: "ifconfig eth2 down"
  end

  # Add Custom Ports From Configuration
  if settings.has_key?("ports")
    settings["ports"].each do |port|
      config.vm.network "forwarded_port",
        guest: port["guest"],
        host: port["host"],
        protocol: port["protocol"] ||= "tcp",
        auto_correct: true
    end
  end

  config.vm.synced_folder "./.vagrant", "/vagrant", nil

  # Register All Of The Configured Shared Folders
  if settings.include? 'folders'
    settings["folders"].each do |folder|
      mount_opts = []

      if (folder["type"] == "nfs")
        mount_opts = folder["mount_opts"] ? folder["mount_opts"] : ['rw', 'vers=3', 'actimeo=1']
      end

      config.vm.synced_folder folder["map"], folder["to"],
        type: folder["type"] ||= nil,
        mount_options: mount_opts
    end
  else
    config.vm.synced_folder ".", "/home/vagrant/apps/#{File.basename(File.dirname(__FILE__))}",
      type: "nfs",
      mount_options: ['rw', 'vers=3', 'actimeo=1']
  end

  # Configure The Public Key For SSH Access
  if settings.include? 'authorize'
    config.vm.provision "shell" do |s|
      s.inline = "echo $1 | grep -xq \"$1\" /home/vagrant/.ssh/authorized_keys || echo $1 | tee -a /home/vagrant/.ssh/authorized_keys"
      s.args = [File.read(File.expand_path(settings["authorize"]))]
    end
  end

  # Copy The SSH Private Keys To The Box
  if settings.include? 'keys'
    settings["keys"].each do |key|
      config.vm.provision "shell" do |s|
        s.privileged = false
        s.inline = "echo \"$1\" > /home/vagrant/.ssh/$2 && chmod 600 /home/vagrant/.ssh/$2"
        s.args = [File.read(File.expand_path(key)), key.split('/').last]
      end
    end
  end

  config.ssh.shell = "bash -c 'BASH_ENV=/etc/profile exec bash'"

  config.vm.provision :shell, :path => ".vagrant/provision/01_pre-provisioning.sh"

  config.vm.provision :puppet do |puppet|
    puppet.facter = { "fqdn" => $host_name, "hostname" => $host_name }
    puppet.manifests_path = ".vagrant/manifests"
    puppet.manifest_file  = "base.pp"
    puppet.module_path = ".vagrant/modules"
  end

  config.vm.provision :shell, :path => ".vagrant/provision/02_post-provisioning.sh"
end
