### logging

Logger based on [zap](https://github.com/uber-go/zap) including middleware for standard http Handler and for [httprouter](https://github.com/julienschmidt/httprouter)

### install

```
go get gitlab.cc.toom.de/dev/producks/zap-logging
```

### usage

```
package main

import "gitlab.cc.toom.de/dev/producks/zap-logging"

func main() {
    logging.Logger.Infof("i want to log number %d", 15)
    logging.Logger.Errorf("i want to log error %v", err)
    logging.Logger.Fatal("logging fatal")
}
```

### logging application start and stop events

```
package main

import "gitlab.cc.toom.de/dev/producks/zap-logging"

func main() {
    // config value can be parsed into LogAppStart()
    logging.LogAppStart("my application name", cfg)

    logging.LogAppStop("my application name", os.Interrupt, nil)
}
```

### logging middleware standard/mux router

```
package some

import (
    "gitlab.cc.toom.de/dev/producks/zap-logging"
    "github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
    router := mux.NewRouter()

    router.Handle("/path", logging.HTTPHandlerMiddleware(MyHandler)).Methods(http.MethodGet)

    return router
}
```

### logging middleware httprouter router

```
package some

import (
    "gitlab.cc.toom.de/dev/producks/zap-logging"
    "github.com/julienschmidt/httprouter"
)

func NewRouter() *httprouter.Router {
    router := httprouter.New()

    router.GET("/path", logging.HTTPRouterMiddleware(MyHandler))

    return router
}
```
