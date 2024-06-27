## Installing

Go **1.18** is required as MoltenChain uses generics.


1. Install golang https://golang.org/doc/install
   1. For Linux
      1. `$ wget https://go.dev/dl/go1.18.1.linux-amd64.tar.gz`
      2. `$ sudo rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.1.linux-amd64.tar.gz`
      3. `$ sudo nano ~/.bashrc`
      4. Add at the bottom of the file
         ```
         export GOPATH=/usr/local/go
         export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
         ```
      5. `$ sudo source ~/.bashrc` 
2. Installing missing packages `go get -t .`
3. Run the node

### Checking and Installing a specific go version
1. `go env GOROOT`
2. download from https://golang.org/doc/manage-install
