This directory contains the following files:

* nginx.tmpl - Should be used to track migrated directives. We will test later to see
if the ending result of it is the same as the one parsed by go-crossplane
* various nginx.conf - Will be used to see if the template parser reacts as 
expected, creating files that matches and can be parsed by go-crossplane

TODO: move files to embed.FS 