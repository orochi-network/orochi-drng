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
	keyType int
	privKey p2pCrypto.PrivKey
	pubKey  p2pCrypto.PubKey
}

// JSON structure
type JSON struct {
	KeyType int    `json:"type"`
	SignKey bool   `json:"signKey"`
	Key     string `json:"key"`
}

// NewEd25519 generates a new Ed25519 KeyPair
func NewEd25519() (*KeyPair, error) {
	return New(p2pCrypto.Ed25519, 256)
}

// NewSecp256k1 generate a new Secp256k1 KeyPair
func NewSecp256k1() (*KeyPair, error) {
	return New(p2pCrypto.Secp256k1, 256)
}

// New generate a new key pair
func New(typ int, bits int) (*KeyPair, error) {
	p, v, err := p2pCrypto.GenerateKeyPairWithReader(typ, bits, rand.Reader)
	if err == nil {
		return &KeyPair{keyType: typ, privKey: p, pubKey: v}, nil
	}
	return nil, err
}

// FromPrivateKey restore a KeyPair from a raw private key
func FromPrivateKey(typ int, b []byte) (*KeyPair, error) {
	var p p2pCrypto.PrivKey
	var err error
	if typ == p2pCrypto.Ed25519 {
		p, err = p2pCrypto.UnmarshalEd25519PrivateKey(b)
	} else if typ == p2pCrypto.Secp256k1 {
		p, err = p2pCrypto.UnmarshalSecp256k1PrivateKey(b)
	}
	if err == nil {
		return &KeyPair{privKey: p, pubKey: p.GetPublic()}, nil
	}
	return nil, err
}

// FromBase64PrivateKey restore a KeyPair from a base 64 private key
func FromBase64PrivateKey(typ int, b string) (*KeyPair, error) {
	p, err := p2pCrypto.ConfigDecodeKey(b)
	if err == nil {
		return FromPrivateKey(typ, p)
	}
	return nil, err
}

// FromPublicKey restore a KeyPair from a raw public key
func FromPublicKey(b []byte) (*KeyPair, error) {
	v, err := p2pCrypto.UnmarshalEd25519PublicKey(b)
	if err == nil {
		return &KeyPair{pubKey: v}, nil
	}
	return nil, err
}

// FromBase64PublicKey restore KeyPair from its public key, this key can not use to sign
func FromBase64PublicKey(b string) (*KeyPair, error) {
	v, err := p2pCrypto.ConfigDecodeKey(b)
	if err == nil {
		return FromPublicKey(v)
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
				return FromBase64PrivateKey(jsonKey.KeyType, jsonKey.Key)
			}
			return FromBase64PublicKey(jsonKey.Key)
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
