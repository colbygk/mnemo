
# mnemo

## Commands

   mnemo <-dv> command

   mnemo login
   mnemo init <-t/--type (go|julia|node|python|ruby|swift)>
             <-f/--frontend (direct|ha-direct|ha-web|web)>
             <-d|--datastore (mysql:...|pg:...|mc:...|redis:...)>
             <-D|--domain (media.mit.edu|...)> [name]
   mnemo clone
   mnemo pull
   mnemo push
   mnemo commit
   mnemo run
   mnemo deploy

### Building

   $ cd <repo_location>/go
   $ export GOPATH=`pwd`
   $ make all
   $ make mnemo
