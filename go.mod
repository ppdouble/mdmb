module github.com/jessepeterson/mdmb

go 1.13

require (
	github.com/go-kit/kit v0.10.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.7.3
	github.com/groob/finalizer v0.0.0-20170707115354-4c2ed49aabda
	github.com/groob/plist v0.0.0-20220217120414-63fa881b19a5
	github.com/jessepeterson/cfgprofiles v0.3.0
	github.com/micromdm/go4 v0.0.0-20210104222236-8a0936d9e451
	github.com/micromdm/scep/v2 v2.1.0
	github.com/pkg/errors v0.8.1
	go.etcd.io/bbolt v1.3.3
	go.mozilla.org/pkcs7 v0.0.0-20210730143726-725912489c62
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect

)

replace github.com/jessepeterson/cfgprofiles v0.3.0 => ../cfgprofiles
