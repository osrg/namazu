#!/bin/sh
set -x
for f in InspectionEndEvents/*.json; do curl --data @$f http://localhost:10000/api/v1; done

