# OpenShift utils

## Introduction
A small batch of helper utils related to OpenShift and the way we use it.

### Extract projectsetups
This utility will loop through all namespaces/projects in OpenShift and extract Projectsetup for groups of projects that fit together. It 
will filter our openshift-specific namespaces.

Grouping will be done based on name suffix, placing all with common base name and different suffixes into same group. Suffixes used for
matching: ci, sit, nasse, brumm, tussi, test, prod-ready, dev

Output can be written to stdout or as files in a directory.

## Author
kristian@fluxconsulting.no
