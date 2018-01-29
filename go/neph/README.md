
# neph

## Commands

   neph <-dv> command

   neph login
   neph init <-t/--type (go|julia|node|python|ruby|swift)>
             <-f/--frontend (direct|ha-direct|ha-web|web)>
             <-d|--datastore (mysql:...|pg:...|mc:...|redis:...)>
             <-D|--domain (media.mit.edu|...)> [name]
   neph clone
   neph pull
   neph push
   neph commit
   neph run
   neph deploy

### Building

   $ cd <repo_location>/go
   $ export GOPATH=`pwd`
   $ make all
   $ make neph
