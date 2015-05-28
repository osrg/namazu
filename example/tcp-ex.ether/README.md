Earthquake Demo (Ethernet Inspector)
===

Run
---

	
	$ ( cd tcp-ex; go build)
	$ cp config_example.json config.json
	$ ./001-start-containers.sh
	$ ./002-set-pipework.sh
	$ ./011-start-switch.sh ### this runs in the foreground
	$ ./021-start-orchestrator.sh ### this runs in the foreground
	$ ./031-start-inspector.sh ### this runs in the foreground

	$ ./501-run-client.sh
	$ sleep 10
	
	$ ./501-run-client.sh
	$ sleep 10

	$ ./901-kill-containers.sh

