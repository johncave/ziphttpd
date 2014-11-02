ziphttpd
========

The Ziphttpd web server is designed for serving static web files as fast as possible.

It achieves this by applying the gzip algorithm to files of any type you chose and storing the zipped files separtely, ready to be served next time someone wants that file.

Any changes to the uncompressed file will be reflected in the compressed version the very next time it is requested, and saved for subsequent requests for that file. You will only need to flush any upstream caches, if you feel you need one.

No cats, hamsters, or bunnies were harmed in the making of this application.
