# fsnotify
## What is this repository?
`go-fsnotify/fsnotify` was the original name of this project. The project organization was renamed from `go-fsnotify` to `fsnotify` some time ago, meaning Github created a silent redirect to the new organization name. This behavior is dangerous since it allows anyone to register the `go-fsnotify` account and create a `fsnotify` repository, which will eventually be pulled by projects that have dependencies to it.
Dependency squatting can remain undetected for a long period of time because Github silently redirects git pulls, reducing breakage.
## My project is not compiling anymore, what can I do?
Change your dependencies from `github.com/go-fsnotify/fsnotify` to `github.com/fsnotify/fsnotify`
