.PHONY: build run dev clean connect check-term

build:
	go build -o sshblog .

run: build
	./sshblog

dev:
	@go build -o sshblog . && \
	./sshblog > /tmp/sshblog.log 2>&1 & \
	sleep 0.5 && \
	ssh -tt localhost -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null; \
	pkill -f "./sshblog" 2>/dev/null || true

clean:
	rm -f sshblog

connect:
	ssh localhost -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null

check-term:
	@echo "TERM in make: $$TERM"
