stage { 'first':
        before => Stage['second'],
        }

stage { 'second':
        before => Stage['third'],
        }

stage { 'third':
        before => Stage['main']
        }

class { "apt_get::update":
        stage  => first,
        }

  class { 'tools':
          stage => second,
          }

    include apt_get::update
    include tools
    include git
    include vim
    include nginx
    include redis
    include nodejs
    include mongodb
    include golang
