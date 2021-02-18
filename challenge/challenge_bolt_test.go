package challenge

import (
	"crypto/x509"
	"io/ioutil"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	challengestore "github.com/micromdm/scep/challenge/bolt"
	"github.com/micromdm/scep/scep"
	scepserver "github.com/micromdm/scep/server"
)

func TestDynamicChallenge(t *testing.T) {
	db, err := openTempBolt("scep-challenge")
	if err != nil {
		t.Fatal(err)
	}

	depot, err := challengestore.NewBoltDepot(db)
	if err != nil {
		t.Fatal(err)
	}

	// use the exported interface
	store := Store(depot)

	// get first challenge
	challengePassword, err := store.SCEPChallenge()
	if err != nil {
		t.Fatal(err)
	}

	if challengePassword == "" {
		t.Error("empty challenge returned")
	}

	// test store API
	valid, err := store.HasChallenge(challengePassword)
	if err != nil {
		t.Fatal(err)
	}
	if valid != true {
		t.Error("challenge just acquired is not valid")
	}
	valid, err = store.HasChallenge(challengePassword)
	if err != nil {
		t.Fatal(err)
	}
	if valid != false {
		t.Error("challenge should not be valid twice")
	}

	// get another challenge
	challengePassword, err = store.SCEPChallenge()
	if err != nil {
		t.Fatal(err)
	}

	if challengePassword == "" {
		t.Error("empty challenge returned")
	}

	// test CSRSigner middleware
	nullSigner := scepserver.CSRSignerFunc(func(*scep.CSRReqMessage) (*x509.Certificate, error) {
		return nil, nil
	})
	mw := NewCSRSignerMiddleware(depot)
	signer := mw(nullSigner)

	csrReq := &scep.CSRReqMessage{
		ChallengePassword: challengePassword,
	}

	_, err = signer.SignCSR(csrReq)
	if err != nil {
		t.Error(err)
	}

	_, err = signer.SignCSR(csrReq)
	if err == nil {
		t.Error("challenge should not be valid twice")
	}

}

func openTempBolt(prefix string) (*bolt.DB, error) {
	f, err := ioutil.TempFile("", prefix+"-")
	if err != nil {
		return nil, err
	}
	f.Close()
	err = os.Remove(f.Name())
	if err != nil {
		return nil, err
	}

	return bolt.Open(f.Name(), 0644, nil)
}
