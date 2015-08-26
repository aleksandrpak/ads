class meteorjs {
  $home = "/home/vagrant"

  file { "${ home }/meteor-installer":
    source => 'puppet:///modules/meteorjs/meteor-installer',
    owner  => 'vagrant',
    group  => 'vagrant',
    mode   => 755
  }

  exec { "install-meteor":
    command  => "su - vagrant ${ home }/meteor-installer",
    provider => 'shell',
    cwd      => $home,
    timeout  => 3600,
    creates  => "${ home }/.meteor",
    require  => [ Package['curl'], File["${ home }/meteor-installer"] ],
    logoutput => true
  }
}
