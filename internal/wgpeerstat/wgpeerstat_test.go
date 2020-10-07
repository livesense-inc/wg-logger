package wgpeerstat

import (
	"bufio"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func stubReadWGDump(_ string) (lines []string, err error) {
	file, err := os.Open("../../test/wg-show-all-dump.txt")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		log.Fatal(err)
		return
	}

	return
}

func TestWGPeerStat_GetPeerStatus(t *testing.T) {
	readWGDump = stubReadWGDump
	peerStats, _ := GetPeerStats("")

	assert.EqualValues(t, "i+VdaJmF7mSlQlDQnEuFbo1JFicB2X054uN0DF5MICA=", peerStats[0].PublicKey)
	assert.EqualValues(t, "239.14.56.78:64515", peerStats[1].Endpoint)
	assert.EqualValues(t, time.Unix(1599229661, 0), peerStats[2].LatestHandshake)
	assert.EqualValues(t, 11158700284, peerStats[3].TransferRX)
	assert.EqualValues(t, 8037532260, peerStats[3].TransferTX)
}
