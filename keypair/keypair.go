package keypair

import (
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"os"

	p2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

// KeyPair structure
type KeyPair struct {
	privKey p2pCrypto.PrivKey
	pubKey  p2pCrypto.PubKey
}

// JSON structure
type JSON struct {
	SignKey bool   `json:"signKey"`
	Key     string `json:"key"`
}

// New define new key pair
func New() (*KeyPair, error) {
	p, v, err := p2pCrypto.GenerateKeyPairWithReader(p2pCrypto.Ed25519, 256, rand.Reader)
	if err == nil {
		return &KeyPair{privKey: p, pubKey: v}, nil
	}
	return nil, err
}

// FromPrivKey restore key pair in base64
func FromPrivKey(b string) (*KeyPair, error) {
	data, err := p2pCrypto.ConfigDecodeKey(b)
	if err == nil {
		p, err := p2pCrypto.UnmarshalEd25519PrivateKey(data)
		if err == nil {
			return &KeyPair{privKey: p, pubKey: p.GetPublic()}, nil
		}
		return nil, err
	}
	return nil, err
}

// FromPubKey restore key pair in base64, this is verify only
func FromPubKey(b string) (*KeyPair, error) {
	data, err := p2pCrypto.ConfigDecodeKey(b)
	if err == nil {
		v, err := p2pCrypto.UnmarshalEd25519PublicKey(data)
		if err == nil {
			return &KeyPair{privKey: nil, pubKey: v}, nil
		}
		return nil, err
	}
	return nil, err
}

// Write json key to file
func writeToJSON(fid *os.File, jsonKey *JSON) (bool, error) {
	encodedJSON, err := json.Marshal(*jsonKey)
	if err == nil {
		writtenBytes, err := fid.Write(encodedJSON)
		if err == nil {
			return writtenBytes > 0, err
		}
		return false, err
	}
	return false, err
}

// SaveToFile save key pair to file
func (k *KeyPair) SaveToFile(fileName string) (bool, error) {
	fid, err := os.Create(fileName)
	if err == nil {
		jsonKey := new(JSON)
		defer fid.Close()
		// Sign able key
		if k.isAbleToSign() {
			key, err := k.privKey.Raw()
			if err == nil {
				jsonKey.SignKey = true
				jsonKey.Key = p2pCrypto.ConfigEncodeKey(key)
				return writeToJSON(fid, jsonKey)
			}
			return false, err
		}
		// Verify only key
		key, err := k.pubKey.Raw()
		if err == nil {
			jsonKey.SignKey = false
			jsonKey.Key = p2pCrypto.ConfigEncodeKey(key)
			return writeToJSON(fid, jsonKey)
		}
		return false, err
	}
	return false, err
}

// LoadFromFile save key pair to file
func LoadFromFile(fileName string) (*KeyPair, error) {
	fileContent, err := ioutil.ReadFile(fileName)
	if err == nil {
		jsonKey := new(JSON)
		err := json.Unmarshal(fileContent, jsonKey)
		if err == nil {
			if jsonKey.SignKey {
				return FromPrivKey(jsonKey.Key)
			}
			return FromPubKey(jsonKey.Key)
		}
		return nil, err
	}
	return nil, err
}

// isAbleToSign with this key pair
func (k *KeyPair) isAbleToSign() bool {
	return k.privKey != nil
}

// GetPrivateKey of this key pair
func (k *KeyPair) GetPrivateKey() p2pCrypto.PrivKey {
	return k.privKey
}

// GetPublicKey of this key pair
func (k *KeyPair) GetPublicKey() p2pCrypto.PubKey {
	return k.pubKey
}

// GetID of this key pair
func (k *KeyPair) GetID() (peer.ID, error) {
	return peer.IDFromPublicKey(k.GetPublicKey())
}

// Sign data
func (k *KeyPair) Sign(data []byte) (signature []byte, err error) {
	return k.privKey.Sign(data)
}

// Verify data
func (k *KeyPair) Verify(data []byte, signature []byte) (bool, error) {
	return k.pubKey.Verify(data, signature)
}
