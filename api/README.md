# :seedling: KuBean API

```
  _          _                                    _ 
 | | ___   _| |__   ___  __ _ _ __     __ _ _ __ (_)
 | |/ / | | | '_ \ / _ \/ _` | '_ \   / _` | '_ \| |
 |   <| |_| | |_) |  __/ (_| | | | | | (_| | |_) | |
 |_|\_\\__,_|_.__/ \___|\__,_|_| |_|  \__,_| .__/|_|
                                           |_|      
```

Schema of the Kubean API types that are served by the Kubernetes API server as custom resources.

## Purpose

This library is the canonical location of the Kubean API definition. Most likely interaction with this repository is as a dependency of client-go, controller-runtime or OpenAPI client.

It is published separately to provide clean dependency and public access for repos depending on it.
