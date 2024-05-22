DURATION_SECONDS=$((5 * 60))
PROFILE_INTERVAL_SECONDS=$((1 * 60))
PROFILE_DURATION_SECONDS=10

function poll() {
	while true; do
		sleep ${PROFILE_INTERVAL_SECONDS}
		curl localhost:6060/debug/pprof/profile?seconds=${PROFILE_DURATION_SECONDS} >profile
	done
}

poll &
echo ${DURATION_SECONDS}
timeout -s INT ${DURATION_SECONDS} ./prof
