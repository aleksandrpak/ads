class nodejs {

    Exec { path => [ "/bin/", "/sbin/" , "/usr/bin/", "/usr/sbin/" ] }

    exec { 'add_node_repo':
        command => 'curl -sL https://deb.nodesource.com/setup_0.12 | sudo -E bash -'
    }

    exec { 'update_node_repo':
        command => '/usr/bin/apt-get update',
        require => Exec['add_node_repo']
    }

    package { 'nodejs':
        ensure => latest,
        require => Exec['update_node_repo'],
    }

}
