class mongodb::pkg {

    Exec { path   => [ "/bin/", "/sbin/" , "/usr/bin/", "/usr/sbin/" ] }

    exec { 'add_mongo_repo_key':
        command   => 'sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10'
    }

    exec { 'add_mongo_repo':
        command   => 'echo "deb http://repo.mongodb.org/apt/ubuntu "$(lsb_release -sc)"/mongodb-org/3.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-3.0.list'
    }

    exec { 'update_mongo_repo':
        command   => '/usr/bin/apt-get update',
        require   => Exec['add_mongo_repo_key', 'add_mongo_repo']
    }

    package { 'mongodb-org':
        ensure    => latest,
        require   => Exec['update_mongo_repo']
    }

    file { 'remove_mongo_data_folder_with_old_db_engine':
        ensure    => 'absent',
        recurse   => true,
        purge     => true,
        force     => true,
        path      => '/var/lib/mongodb',
        require   => Package['mongodb-org']
    }
}

class mongodb::conf {

    Exec { path   => [ "/bin/", "/sbin/" , "/usr/bin/", "/usr/sbin/" ] }

    File {
        require   => Class['mongodb::pkg'],
        notify    => Class['mongodb::service'],
        owner     => 'root',
        group     => 'root',
        mode      => 644
    }

    file { 'mongodb-config':
        source    => 'puppet:///modules/mongodb/mongod.conf',
        path      => '/etc/mongod.conf'
    }
}

class mongodb::service {
    service { 'mongod':
        ensure    => running,
        enable    => true,
        require   => Class['mongodb::conf'],
    }
}

class mongodb {
    include mongodb::pkg, mongodb::conf, mongodb::service
}

