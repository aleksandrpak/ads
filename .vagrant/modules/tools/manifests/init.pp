class tools {
    $packages = [
        'language-pack-ru',
        'python-software-properties',
        'curl',
        'imagemagick',
        'build-essential',
        'python',
        'htop',
        'openssl',
        'libssl-dev',
        'cachefilesd'
    ]

    package { $packages:
        ensure => installed,
    }
}
