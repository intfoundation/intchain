package ethapi

import (
	"fmt"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/common/math"
	"github.com/intfoundation/intchain/crypto"
	"testing"
	"time"
)

var FromAddr = "INT3MccCA7EtMzijJa2zjxoiSYzbNLE4"
var PubKey = "0x618CEAF6AD449B826E2521222A94426B82800202332251F0929EC47B36A647C65E00D2EA34C07A8EF7953C2E1555D8321449423CCFB0B64BB13090E7A433114D68F1C1891BAA20101E5CC8E2B10E207F5D21D1A1116547E1EED5E92FDFE4F5E58119C5267B82AE06BBA5016827396B74E1ECDCC3801746242CA24C7749EB2F88"
var Amount = "0x152d02c7e14af68000000"
var Salt = "like"
var VoteHash1 = "0xc6335e23dd8ba330b2d3c34acdeb2dfd0b07d30dfc2d5f9ca1b0d62e147788f0" // false
var VoteHash2 = "0xa431ab9cb5d2750faeed74945d10c69372b938c2470d5b140de29f4d4aa22025" // true
var VoteHash3 = "0xb2aa67b3cf56dcb41097d72024962c03d4fba2a9892cc37e348243b85bf58c27" // false

func TestVoteHash(t *testing.T) {
	pubKey, err := hexutil.Decode(PubKey)
	if err != nil {
		t.Fatal(err)
	}
	byteData := [][]byte{
		[]byte(FromAddr),
		pubKey,
		common.LeftPadBytes(math.MustParseBig256("1600000000000000000000000").Bytes(), 1),
		[]byte(Salt),
	}

	hash := crypto.Keccak256Hash(concatCopyPreAllocate(byteData))
	fmt.Printf("vote hash %v\n", hash.String())

}

func TestGoTime(t *testing.T) {
	nowTime := time.Now().Unix()
	fmt.Printf("now %v\n", nowTime)

	d := 24 * time.Hour
	fmt.Printf("duration %v\n", d)
	fmt.Printf("duration string %v\n", d.String())
	fmt.Printf("duration seconds %v\n", d.Seconds())
}
