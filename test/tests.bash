
trap "make test_cleanup" INT TERM EXIT

export GOSHE_DEBUG=1

T_06runbinary() {
  result="$(bin/goshe)"
}
