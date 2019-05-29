#!/bin/bash -eu
# Ambassador-compatible wrapper script for composer. Expects the environment
# variable $WF_JSON to be set to a workflow value.  Other flags could also be
# set, but not -indir and -outdir.
COMPOSER=${COMPOSER:-composer}
MIGRATE=${MIGRATE:-migrate}

# Well-known directory name set by the workload service.
# (See https://github.com/Synthace/microservice/tree/master/cmd/workload#compatible-containers )
DATA_DIR=${DATA_DIR:-/data}

# (See https://github.com/Synthace/microservice/tree/master/cmd/ambassador )
{ while ! [[ -r $DATA_DIR/inputReady ]]; do sleep 0.1; done; }
dd if=$DATA_DIR/inputReady of=/dev/null bs=1 count=1
trap "{ dd if=/dev/zero of=$DATA_DIR/outputReady bs=1 count=1; }" EXIT

find $DATA_DIR -type f
$MIGRATE -from=${DATA_DIR}/input/workflow/request.pb -outdir=${DATA_DIR}/scratch -format=protobuf - <<<$WF_JSON

cp -a ${DATA_DIR}/scratch/* ${DATA_DIR}/input

find $DATA_DIR -type f
$COMPOSER -indir=${DATA_DIR}/input -outdir=${DATA_DIR}/output
find $DATA_DIR -type f
