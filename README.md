paas-db-admin-boshrelase
------------------------

Basic release to perform administration operations on databases (e.g. RDS) used
by the VMs of a Bosh deployment.

This intends to solve the problem of how to initialise external services like
RDS to be used by VMs. For instance, we need to create some databases and roles
on a PostgreSQL database to be used by different services (e.g. UAA) before we
deploy the VM, but the RDS is accesible only from the same VPC and there is no
previous VMs running in the VPC.

job: init-db
------------

Its intention is to create databases and roles in `pre-start` script.  This job
can be colocated with the other services using these databases, and you can let
them race. The other jobs might fail starting, but Monit would restart them
immediately after.

See the `jobs/init-db/spec` file for more details.
