# 1.1.0
  - Enabled locking for the loggers list to avoid problems when using the
    package concurrently.
  - Some refactoring and added more testing, benchmarking.
  - Added `ImportLoggers` which will repopulate the internal logger list with
    the given list.
  - Added `Trace` Priority which is lower than `Debug`.
  - Now allowing arrays of string when constructiong a logger which will use the
    default sepperator to save the name of the logger.
  - Will now save logger level when we got the parent logger to increase
    performance. Can be disabled by setting `SaveLoggerLevels` to `false`.

# 1.0.0
First Release
