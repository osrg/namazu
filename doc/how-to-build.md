# How to Build Earthquake
You need to install golang 1.5 or later to build libearthquake.so.

    $ git clone -b release-branch.go1.4 https://go.googlesource.com/go $HOME/go1.4
    $ (cd $HOME/go1.4/src; ./make.bash) # golang 1.4 (in $HOME/go1.4/bin) is required to build golang 1.5
    $ git clone https://go.googlesource.com/go $HOME/go
    $ (cd $HOME/go/src; ./make.bash)
    $ export PATH=$HOME/go/bin:$PATH 

Then, please just run `./build` in the top directory of Earthquake.

