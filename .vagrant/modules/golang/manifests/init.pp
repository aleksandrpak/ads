class golang (
  $version = "1.5",
  $homedir = "/home/vagrant"
) {

  Exec { path   => [ "/bin/", "/sbin/" , "/usr/bin/", "/usr/sbin/", "/usr/local/go/bin" ] }

  $goTarPath = "/usr/local/src/go$version.linux-amd64.tar.gz"

  exec { "download-golang":
    command => "wget --no-check-certificate -O $goTarPath https://storage.googleapis.com/golang/go$version.linux-amd64.tar.gz",
    creates => "$goTarPath"
  }

  exec { "unarchive-golang":
    command => "tar -C /usr/local -xzf $goTarPath",
    require => Exec["download-golang"]
  }

  exec { "remove-previous":
    command => "rm -r /usr/local/go",
    onlyif  => [
      "test -d /usr/local/go",
      "which go && test `go version | cut -d' ' -f 3` != 'go$version'",
    ],
    before  => Exec["unarchive-golang"],
  }

  file { "/etc/profile.d/golang.sh":
    content => template("golang/golang.sh.erb"),
    owner   => root,
    group   => root,
    mode    => "a+x",
  }

  if ! defined(Package["git"]) {
    package { "git":
      ensure => present,
    }
  }

  if ! defined(Package["bzr"]) {
    package { "bzr":
      ensure => present,
    }
  }
}
