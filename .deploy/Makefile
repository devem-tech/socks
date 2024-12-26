.PHONY: up
up:
	docker pull ghcr.io/devem-tech/socks:main
	docker stop socks || true
	docker rm socks || true
	docker run -d \
	  -p 7010:7010 \
	  --name socks \
	  --network net \
	  --restart always \
	  --env-file /opt/containers/socks/.env \
	  ghcr.io/devem-tech/socks:main