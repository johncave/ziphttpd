ziphttpd
========

The Ziphttpd (zippy) web server is designed for serving static web files as fast as possible.

It achieves this by applying the gzip algorithm to files of any type you choose and storing the zipped files separately, ready to be served next time someone wants that file.

Any changes to the uncompressed file are reflected in the compressed version the very next time it is requested, and saved for subsequent requests for that file. You will only need to flush any upstream caches (if you feel you need one after seeing what ziphttpd can do on its own).

No cats, hamsters, or bunnies were harmed in the making of this application. It is scheduled for beta release in two weeks.
