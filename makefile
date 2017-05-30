# old host
host = ubuntu@ec2-35-163-107-72.us-west-2.compute.amazonaws.com
	
go:
	go build
	rsync -avzrhSP -e "ssh -i spacecafe.pem" spacecafe $(host):/home/ubuntu/stage/

database:		
	rsync -avzrhSP -e "ssh -i spacecafe.pem" data/migrations $(host):/home/ubuntu/stage/data
	rsync -avzrhSP -e "ssh -i spacecafe.pem" data/dbconf.yml $(host):/home/ubuntu/stage/data
	rsync -avzrhSP -e "ssh -i spacecafe.pem" data/redis/redis.conf $(host):/home/ubuntu/stage/data/redis
	# rsync -avzrhSP -e "ssh -i spacecafe.pem" db/redis/redis.conf root@107.170.99.115:/usr/local/redis

static:	
	rsync -avzrhSP -e "ssh -i spacecafe.pem" --exclude='*.go' templates $(host):/home/ubuntu/stage
	rsync -avzrhSP -e "ssh -i spacecafe.pem" static/css $(host):/home/ubuntu/stage/static
	rsync -avzrhSP -e "ssh -i spacecafe.pem" static/js $(host):/home/ubuntu/stage/static
	
deploy:	
	# goose -path data -env production up