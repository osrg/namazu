# A Template for Your Own Test Scenario

## Try Filesystem Inspector
On Terminal 1,
    
    $ earthquake init --force config.toml materials /tmp/template
    $ EQ_DEBUG=1 earthquake run /tmp/template

On Terminal 2,

    $ EQ_DEBUG=1 ~/bin/earthquake inspectors fs -original-dir ~/tmp -mount-point ~/mnt

On Terminal 3,

    $ cat ~/mnt/foobar



## Write Your Own Explorepolicy

    $ emacs mypolicy.go
    $ go build -o mypolicy mypolicy.go

`mypolicy` provides CLI, same as `earthquake`.

    $ ./mypolicy init --force config_mypolicy.go materials /tmp/mypolicy
    $ ./mypolicy run /tmp/mypolicy

...
