# A Template for Your Own Test Scenario

## Try Filesystem Inspector
On Terminal 1,
    
    $ nmz init --force config.toml materials /tmp/template
    $ NMZ_DEBUG=1 nmz run /tmp/template

On Terminal 2,

    $ NMZ_DEBUG=1 ~/bin/nmz inspectors fs -original-dir ~/tmp -mount-point ~/mnt

On Terminal 3,

    $ cat ~/mnt/foobar



## Write Your Own Explorepolicy

    $ emacs mypolicy.go
    $ go build -o mypolicy mypolicy.go

`mypolicy` provides CLI, same as `nmz`.

    $ ./mypolicy init --force config_mypolicy.go materials /tmp/mypolicy
    $ ./mypolicy run /tmp/mypolicy

...
