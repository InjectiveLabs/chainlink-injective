package ocrkey

import (
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"golang.org/x/crypto/curve25519"
)

type OCR2KeyWrapper struct {
	ocr KeyV2
}

func NewOCR2KeyWrapper(key KeyV2) OCR2KeyWrapper {
	return OCR2KeyWrapper{key}
}

func (o OCR2KeyWrapper) OffchainSign(msg []byte) (signature []byte, err error) {
	return o.ocr.SignOffChain(msg)
}

func (o OCR2KeyWrapper) ConfigDiffieHellman(point [curve25519.PointSize]byte) (sharedPoint [curve25519.PointSize]byte, err error) {
	out, err := o.ocr.ConfigDiffieHellman(&point)
	return *out, err
}

func (o OCR2KeyWrapper) OffchainPublicKey() types.OffchainPublicKey {
	return types.OffchainPublicKey(o.ocr.PublicKeyOffChain())
}

func (o OCR2KeyWrapper) ConfigEncryptionPublicKey() types.ConfigEncryptionPublicKey {
	return o.ocr.PublicKeyConfig()
}
