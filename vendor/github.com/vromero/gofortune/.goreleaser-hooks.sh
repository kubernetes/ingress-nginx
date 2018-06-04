#!/bin/sh

for DIR in $(find dist -name gofortune -exec dirname \{\} \; | grep -v Windows); do cd $DIR && ln -s gofortune fortune; ln -s gofortune strfile; cd -; done

