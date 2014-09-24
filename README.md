go-pool
=======

This package contains a static and a dynamic pool which have:

* Standard interface
* Timeout for `Get()`
* Callbacks for `Get()` and `Put()`
* Persistence (items are _not_ garbage-collected)

The static pool grows to its size after N many calls to `Get()`, where N equals the pool size. This is useful for cases where the number of items is also static, for example a worker pool where N many workers are always used.

The dynamic pool grows to its size if the rate of calls to `Get()` exceeds the rate of calls to `Put()`. Else, the pool size can, for example, remain at one if there's a single caller (in which case the same item is reused repeatedly).

Neither pool implements pruning or draining.

These pools were inspired by http://openmymind.net/Introduction-To-Go-Channels/, the built-in sync.Pool, and https://github.com/go-sql-driver/mysql.
