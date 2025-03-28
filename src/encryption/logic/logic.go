package logic

import (
	"AnonyMouse/encryption/model"
	"encoding/hex"
	"encoding/json"
	"syscall/js"
)

func GenOnRegister(this js.Value, args []js.Value) interface{} {
	var result model.RegPack
	var err error

	result.IK, result.SK, result.Sig, err = genOnRegister()
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	res, _ := json.Marshal(result)
	return string(res)
}

// Paramters (7): other_pub_IdentityKey string, other_pub_SchnorrKey string, other_SchnorrSignature string,
//	other_pub_OnetimePreKey string, my_priv_identityKey string, plainText string, timestamp string
func X3DHSender(this js.Value, args []js.Value) interface{} {
	var result model.X3DHSendPack

	if len(args) != 7 {
		result.Err = "Invalid number of args"
		res, _ := json.Marshal(result)
		return string(res)
	}

	IKB, err := hex.DecodeString(args[0].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	SKB, err := hex.DecodeString(args[1].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	SigB, err := hex.DecodeString(args[2].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	OPKB, err := hex.DecodeString(args[3].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	ikA, err := hex.DecodeString(args[4].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	plainText := []byte(args[5].String())
	timestamp := []byte(args[6].String())

	valid, err := schnorrVerify(IKB, SKB, SigB)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	} else if !valid {
		result.Err = "Signature Is Not Valid"
		res, _ := json.Marshal(result)
		return string(res)
	}

	rK, eK, err := x3dhSender(IKB, SKB, OPKB, ikA, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rK, sCK, err := ratchetNext(rK, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	sCK, mK, err := ratchetNext(sCK, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	cipherText, err := encryptGCM(mK, plainText)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	dhK, err := generateKeyPair()
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	eKStr := hex.EncodeToString(eK)
	rKStr := hex.EncodeToString(rK)
	sCKStr := hex.EncodeToString(sCK)
	mKStr := hex.EncodeToString(mK)

	result.CipherText, result.EK, result.RK, result.SCK, result.MK, result.DHK = cipherText, eKStr, rKStr, sCKStr, mKStr, dhK

	res, _ := json.Marshal(result)
	return string(res)
}

// Paramters (4): other_pub_DiffieHellmanKey string, rootKey string, plainText string, timestamp string
func SenderFirst(this js.Value, args []js.Value) interface{} {
	var result model.SendFirstPack

	if len(args) != 4 {
		result.Err = "Invalid number of args"
		res, _ := json.Marshal(result)
		return string(res)
	}

	dhKB, err := hex.DecodeString(args[0].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rK, err := hex.DecodeString(args[1].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	plainText := []byte(args[2].String())
	timestamp := []byte(args[3].String())

	dhSK, dhKA, err := ecdhSend(dhKB)

	rK, sCK, err := ratchetNext(append(rK, dhSK...), timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	sCK, mK, err := ratchetNext(sCK, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	cipherText, err := encryptGCM(mK, plainText)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rKStr := hex.EncodeToString(rK)
	sCKStr := hex.EncodeToString(sCK)
	mKStr := hex.EncodeToString(mK)

	result.CipherText, result.RK, result.SCK, result.MK, result.DHK = cipherText, rKStr, sCKStr, mKStr, dhKA

	res, _ := json.Marshal(result)
	return string(res)
}

// Paramters (3): sendingChainKey string, plainText string, timestamp string
func Sender(this js.Value, args []js.Value) interface{} {
	var result model.SendPack

	if len(args) != 3 {
		result.Err = "Invalid number of args"
		res, _ := json.Marshal(result)
		return string(res)
	}

	sCK, err := hex.DecodeString(args[0].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	plainText := []byte(args[1].String())
	timestamp := []byte(args[2].String())

	sCK, mK, err := ratchetNext(sCK, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	cipherText, err := encryptGCM(mK, plainText)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	sCKStr := hex.EncodeToString(sCK)
	mKStr := hex.EncodeToString(mK)

	result.CipherText, result.SCK, result.MK = cipherText, sCKStr, mKStr

	res, _ := json.Marshal(result)
	return string(res)
}

// Paramters (9): other_pub_IdentityKey string, other_pub_SchnorrKey string, other_SchnorrSignature string,
//	other_pub_EphamiralKey string, my_priv_IdentityKey string, my_priv_SchnorrKey string,
//	my_priv_OnetimePreKey string, cipherText string, timestamp string
func X3DHReceiver(this js.Value, args []js.Value) interface{} {
	var result model.X3DHRecPack

	if len(args) != 9 {
		result.Err = "Invalid number of args"
		res, _ := json.Marshal(result)
		return string(res)
	}

	IKB, err := hex.DecodeString(args[0].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	SKB, err := hex.DecodeString(args[1].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	SigB, err := hex.DecodeString(args[2].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	EKB, err := hex.DecodeString(args[3].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	ikA, err := hex.DecodeString(args[4].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	skA, err := hex.DecodeString(args[5].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	opkA, err := hex.DecodeString(args[6].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	cipherText, err := hex.DecodeString(args[7].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	timestamp := []byte(args[8].String())

	valid, err := schnorrVerify(IKB, SKB, SigB)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	} else if !valid {
		result.Err = "Signature Is Not Valid"
		res, _ := json.Marshal(result)
		return string(res)
	}

	rK, err := x3dhReceiver(IKB, EKB, ikA, skA, opkA, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rK, rCK, err := ratchetNext(rK, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rCK, mK, err := ratchetNext(rCK, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	plainText, err := decryptGCM(mK, cipherText)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rKStr := hex.EncodeToString(rK)
	rCKStr := hex.EncodeToString(rCK)

	result.PlainText, result.RK, result.RCK = plainText, rKStr, rCKStr

	res, _ := json.Marshal(result)
	return string(res)
}

// Paramters (5): other_pub_DiffieHellmanKey string, my_priv_DiffieHellmanKey string, rootKey string, cipherText string, timestamp string
func ReceiverFirst(this js.Value, args []js.Value) interface{} {
	var result model.RecFirstPack

	if len(args) != 5 {
		result.Err = "Invalid number of args"
		res, _ := json.Marshal(result)
		return string(res)
	}

	dhKB, err := hex.DecodeString(args[0].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}
	dhKA, err := hex.DecodeString(args[1].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}
	rK, err := hex.DecodeString(args[2].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}
	cipherText, err := hex.DecodeString(args[3].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	timestamp := []byte(args[4].String())

	dhSK, err := ecdhRec(dhKA, dhKB)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rK, rCK, err := ratchetNext(append(rK, dhSK...), timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rCK, mK, err := ratchetNext(rCK, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	plainText, err := decryptGCM(mK, cipherText)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rKStr := hex.EncodeToString(rK)
	rCKStr := hex.EncodeToString(rCK)

	result.PlainText, result.RK, result.RCK = plainText, rKStr, rCKStr

	res, _ := json.Marshal(result)
	return string(res)
}

// Paramters (3): receivingChainKey string, cipherText string, timestamp string
func Receiver(this js.Value, args []js.Value) interface{} {
	var result model.RecPack

	if len(args) != 3 {
		result.Err = "Invalid number of args"
		res, _ := json.Marshal(result)
		return string(res)
	}

	rCK, err := hex.DecodeString(args[0].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}
	cipherText, err := hex.DecodeString(args[1].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	timestamp := []byte(args[2].String())

	rCK, mK, err := ratchetNext(rCK, timestamp)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	plainText, err := decryptGCM(mK, cipherText)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	rCKStr := hex.EncodeToString(rCK)

	result.PlainText, result.RCK = plainText, rCKStr

	res, _ := json.Marshal(result)
	return string(res)
}

// Paramters (2): messageKey string, cipherText string
func MKDecrypt(this js.Value, args []js.Value) interface{} {
	var result model.DecMSG

	if len(args) != 2 {
		result.Err = "Invalid number of args"
		res, _ := json.Marshal(result)
		return string(res)
	}

	mK, err := hex.DecodeString(args[0].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}
	cipherText, err := hex.DecodeString(args[1].String())
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}

	plainText, err := decryptGCM(mK, cipherText)
	if err != nil {
		result.Err = err.Error()
		res, _ := json.Marshal(result)
		return string(res)
	}
	
	result.PlainText = plainText

	res, _ := json.Marshal(result)
	return string(res)
}

// Parameters (1): numKeys int
func GenerateDHKeys(this js.Value, args []js.Value) interface{} {
	var result model.KeysGenRes

	if len(args) < 1 {
		result.Err = "Invalid number of args"
		res, _ := json.Marshal(result)
		return string(res)
	}

	numKeys := args[0].Int()

	for i := 0; i < numKeys; i++ {
		key, err := generateKeyPair()
		if err != nil {
			result.Err = err.Error()
			res, _ := json.Marshal(result)
			return string(res)
		}
		result.Keys = append(result.Keys, key)
	}

	res, _ := json.Marshal(result)
	return string(res)
}
