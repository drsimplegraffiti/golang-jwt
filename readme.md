### install make

```bash
choco install make
```

#### Redis

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install redis-server -y
sudo systemctl enable redis-server #enable redis as a service
sudo systemctl start redis-server # start the redis-server
sudo systemctl status redis-server #check status
redis-cli ping #ping redis
```

### Allow remote connection

```bash
sudo nano /etc/redis/redis.conf

# Find the line:
bind 127.0.0.1 ::1

# Change it to:
bind 0.0.0.0
# Also set protected-mode no.
# then restart :
sudo systemctl restart redis-server

# Now Redis will be available at:
localhost:6379
```
