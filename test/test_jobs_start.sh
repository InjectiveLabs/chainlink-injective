#!/bin/bash

set -e

./test/scripts/cl_start_job.sh job_linkusdc_ei0 ./test/jobs/job_linkusdc_ei0.toml
./test/scripts/cl_start_job.sh job_linkusdc_ei1 ./test/jobs/job_linkusdc_ei1.toml
./test/scripts/cl_start_job.sh job_linkusdc_ei2 ./test/jobs/job_linkusdc_ei2.toml
./test/scripts/cl_start_job.sh job_linkusdc_ei3 ./test/jobs/job_linkusdc_ei3.toml
