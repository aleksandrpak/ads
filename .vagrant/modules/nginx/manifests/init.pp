class nginx::pkg {
    Exec { path => [ "/bin/", "/sbin/" , "/usr/bin/", "/usr/sbin/" ] }

    exec { 'add_nginx_repo_key':
        command => 'wget -O - http://nginx.org/keys/nginx_signing.key | sudo apt-key add -'
    }

    exec { 'add_nginx_repo':
        command => 'echo "deb http://nginx.org/packages/ubuntu/ "$(lsb_release -sc)" nginx" | sudo tee /etc/apt/sources.list.d/nginx.list'
    }

    exec { 'update_nginx_repo':
        command => '/usr/bin/apt-get update',
        # refreshonly => true,
        require => Exec['add_nginx_repo_key', 'add_nginx_repo']
    }

    package { 'nginx':
        ensure  => latest,
        require => Exec['update_nginx_repo']
    }
}

class nginx::conf {
    File {
        require => Class['nginx::pkg'],
        notify => Class['nginx::service'],
        owner => 'root',
        group => 'root',
        mode => 644
    }

    file { '/etc/nginx/nginx.conf':
        source => 'puppet:///modules/nginx/nginx.conf'
    }


    file { '/usr/share/nginx/certs':
        path => '/usr/share/nginx/certs',
        ensure => directory,
        owner => root,
        group => root,
        purge => true,
        recurse => true,
        source => "puppet:///modules/nginx/certs",
    }


    file { 'nginx/sites-available':
        path => '/etc/nginx/sites-available',
        ensure => directory,
        owner => root,
        group => root,
        purge => true,
        recurse => true,
        source => "puppet:///modules/nginx/sites-available",
    }

    file { 'nginx/sites-enabled':
        path => '/etc/nginx/sites-enabled',
        ensure => directory,
        owner => root,
        group => root,
        purge => true,
        recurse => true,
        source => "puppet:///modules/nginx/sites-enabled",
    }

    file { '/etc/nginx/sites-enabled/vhosts.dev':
        ensure => 'link',
        target => '/etc/nginx/sites-available/vhosts.dev'
    }

    file { 'var/www':
        path => '/var/www',
        ensure => directory,
        owner => vagrant,
        group => vagrant,
    }


}

class nginx::service {
    service { 'nginx':
        ensure => running,
        enable => true,
        require => Class['nginx::conf'],
    }
}

class nginx {
    include nginx::pkg, nginx::conf, nginx::service
}
