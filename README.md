# AWS MySQL Through Bastion in Go
Example Go program that connects to a MySQL AWS/RDS instance through a bastion AMI
I needed to be able to run queries against a MySQL database that is hosted using Amazon's
RDS system that is accessable only through a bastion machine acting as a firewall to the VPC.

I found a number of good hints online for reverse proxying, etc., but nothing that actually
worked for me. This particular program *does* work in that it connects and returns results, which
is what really matters, so I thought I'd clean it up and share in case it helps anyone else.

Note that this program assumes you're working with a pem file.

