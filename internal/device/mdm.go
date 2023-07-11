package device

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/groob/plist"
	"go.mozilla.org/pkcs7"
)

func (c *MDMClient) authenticate() error {
	log.Println("authenticate")
	ar := &AuthenticationRequest{
		DeviceName:  c.Device.ComputerName,
		MessageType: "Authenticate",
		Topic:       c.MDMPayload.Topic,
		UDID:        c.Device.UDID,

		// non-required fields
		SerialNumber: c.Device.Serial,
	}

	return c.checkinRequest(ar)
}

// AuthenticationRequest ...
type AuthenticationRequest struct {
	BuildVersion string `plist:",omitempty"`
	DeviceName   string
	IMEI         string `plist:",omitempty"`
	MEID         string `plist:",omitempty"`
	MessageType  string
	Model        string `plist:",omitempty"`
	ModelName    string `plist:",omitempty"`
	OSVersion    string `plist:",omitempty"`
	ProductName  string `plist:",omitempty"`
	SerialNumber string `plist:",omitempty"`
	Topic        string
	UDID         string
	EnrollmentID string `plist:",omitempty"` // macOS 10.15 and iOS 13.0 and later
}

type ErrorChain struct {
	ErrorCode            int
	ErrorDomain          string
	LocalizedDescription string
	USEnglishDescription string `plist:",omitempty"`
}

type ConnectRequest struct {
	UDID        string
	CommandUUID string `plist:",omitempty"`
	Status      string
	ErrorChain  []ErrorChain `plist:",omitempty"`

	RequestType string `plist:",omitempty"`
}

// type ConnectResponse struct {
// 	Command     interface{}
// 	CommandUUID string
// }

func (c *MDMClient) MdmSignature(body []byte) (string, error) {
	return c.mdmP7Sign(body)
}

// Generates "SignMessage" HTTP header data
func (c *MDMClient) mdmP7Sign(body []byte) (string, error) {
	if c.IdentityCertificate == nil || c.IdentityPrivateKey == nil {
		return "", errors.New("device identity invalid")
	}
	signedData, err := pkcs7.NewSignedData(body)
	if err != nil {
		return "", err
	}
	signedData.AddSigner(c.IdentityCertificate, c.IdentityPrivateKey, pkcs7.SignerInfoConfig{})
	signedData.Detach()
	sig, err := signedData.Finish()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

type TokenUpdateRequest struct {
	AwaitingConfiguration bool   `plist:",omitempty"`
	EnrollmentID          string `plist:",omitempty"` // macOS 10.15 and iOS 13.0 and later
	EnrollmentUserID      string `plist:",omitempty"` // macOS 10.15 and iOS 13.0 and later
	MessageType           string
	NotOnConsole          bool `plist:",omitempty"`
	PushMagic             string
	Token                 []byte
	Topic                 string
	UDID                  string
	UnlockToken           []byte `plist:",omitempty"`
	UserShortName         string `plist:",omitempty"`
	UserID                string `plist:",omitempty"`
	UserLongName          string `plist:",omitempty"`
}

func (c *MDMClient) checkinRequest(i interface{}) error {
	plistBytes, err := plist.Marshal(i)
	if err != nil {
		return err
	}

	mdmSig, err := c.mdmP7Sign(plistBytes)
	if err != nil {
		return err
	}

	ciURL := c.MDMPayload.CheckInURL
	if ciURL == "" {
		ciURL = c.MDMPayload.ServerURL
	}

	client := &http.Client{}
	req, err := http.NewRequest("PUT", ciURL, bytes.NewReader(plistBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Mdm-Signature", mdmSig)
	req.Header.Set("Content-Type", "application/x-apple-aspen-mdm-checkin")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("Checkin request failed with HTTP status: %d", res.StatusCode)
	}

	return nil
}

//func (c *MDMClient) TokenUpdate(addl string) error {
//	log.Println("TokenUpdate")
//	tu := &TokenUpdateRequest{
//		MessageType: "TokenUpdate",
//		PushMagic:   "fakePushMagic" + addl,
//		Token:       []byte("fakeToken" + addl),
//		Topic:       c.MDMPayload.Topic,
//		UDID:        c.Device.UDID,
//	}
//	return c.checkinRequest(tu)
//}

//import "encoding/hex"
//
//// IsDeviceTokenValid checks if s is a hexadecimal token of the correct length.
//func IsDeviceTokenValid(s string) bool {
//	// TODO: In 2016, they may be growing up to 100 bytes (200 hexadecimal digits).
//	if len(s) < 64 || len(s) > 200 {
//		return false
//	}
//	_, err := hex.DecodeString(s)
//	return err == nil
//}


func (c *MDMClient) TokenUpdate(addl string, udid string) error {
	pushMagic := strings.ToUpper(uuid.NewString())

	log.Println("TokenUpdate")
	log.Printf("udid %s\n", udid)
	//base64UDID := base64.StdEncoding.EncodeToString([]byte(udid))
	//log.Printf("base64 udid %s\n", base64UDID)
	budid := []byte(udid)
	log.Printf("budid %s\n", budid)
	//ebudid := make([]byte, hex.EncodedLen(len(budid)))
	//hex.Encode(ebudid, budid)
	//log.Printf("hex.EncodeToString(ebudid), wrong, double encoding: %s\n", hex.EncodeToString(ebudid))
	//log.Printf("ebudid, right: %s\n", ebudid)
	//log.Printf("hex.EncodeToString(ebudid), wrong, double encoding: %s\n", hex.EncodeToString(ebudid))
	log.Printf("hex.EncodeToString(budid): %s\n", hex.EncodeToString(budid))


	tu := &TokenUpdateRequest{
		MessageType: "TokenUpdate",
		PushMagic:   pushMagic,
		Token:       budid,
		Topic:       c.MDMPayload.Topic,
		UDID:        c.Device.UDID,
	}
	return c.checkinRequest(tu)
}

type ConnectResponseCommand struct {
	RequestType string
}

type ConnectResponse struct {
	Command     ConnectResponseCommand
	CommandUUID string
}

func (c *MDMClient) Connect() error {
	req := &ConnectRequest{
		UDID:   c.Device.UDID,
		Status: "Idle",
	}
	client := &http.Client{}
	log.Println("The first connect() calling will send Idle status of the mocked device")
	return c.connect(client, req)
}

func httpRequestBytes(client *http.Client, req *http.Request) (bytes []byte, res *http.Response, err error) {
	log.Println("send request to MDM Server")
	res, err = client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	log.Println("read response")
	bytes, err = ioutil.ReadAll(res.Body)
	return
}

func (c *MDMClient) connect(client *http.Client, connReq interface{}) error {
	log.Println("call connect(), mdm.Connect webhook request")
	log.Println("call connect() as mdm.Connect webhook in loop, until all queued commands on MDM Server were all consumed")
	log.Printf("connReq: %s", connReq)

	if !c.enrolled() {
		return errors.New("device not enrolled")
	}

	log.Println("plist.Marshal to wrap the request to plist")
	plistBytes, err := plist.Marshal(connReq)
	log.Printf("plistBytes %s", plistBytes)
	if err != nil {
		return err
	}

	log.Println("sign the plist")
	mdmSig, err := c.mdmP7Sign(plistBytes)
	if err != nil {
		return err
	}

	log.Println("http.NewRequest to MDM Server to send webhook information")
	req, err := http.NewRequest("PUT", c.MDMPayload.ServerURL, bytes.NewReader(plistBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Mdm-Signature", mdmSig)

	respBytes, res, err := httpRequestBytes(client, req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("Connect Request failed with HTTP status: %d", res.StatusCode)
	}

	if len(respBytes) == 0 {
		log.Println("No command or other information from the response of webhook request from MDM Server. End the loop")
		return nil
	}

	resp := &ConnectResponse{}
	err = plist.Unmarshal(respBytes, &resp)
	log.Printf("respBytes %s", respBytes)
	log.Printf("plist.Unmarshal to unfold the response: %s", resp)
	if err != nil {
		return err
	}

	log.Println("handle the command(if exists) which responded from webhook request")
	nextConnReq, err := c.handleMDMCommand(resp.Command.RequestType, resp.CommandUUID, respBytes)
	log.Printf("resp.Command.RequestType: %s", resp.Command.RequestType)
	log.Printf("resp.CommandUUID: %s", resp.CommandUUID)
	log.Printf("nextConnReq: %s", nextConnReq)
	if err != nil {
		log.Println(err)
		nextConnReq = &ConnectRequest{
			UDID:        c.Device.UDID,
			CommandUUID: resp.CommandUUID,
			RequestType: resp.Command.RequestType,
			Status:      "Error",
			ErrorChain: []ErrorChain{
				{
					ErrorCode:            99998,
					ErrorDomain:          "mdmb-handle-mdm-command",
					LocalizedDescription: "Error handling MDM command",
				},
			},
		}
	}

	if nextConnReq == nil {
		fmt.Println("empty response from handling MDM command")
		nextConnReq = &ConnectRequest{
			UDID:        c.Device.UDID,
			CommandUUID: resp.CommandUUID,
			RequestType: resp.Command.RequestType,
			Status:      "Error",
			ErrorChain: []ErrorChain{
				{
					ErrorCode:            99999,
					ErrorDomain:          "mdmb-handle-mdm-command",
					LocalizedDescription: "Empty response from hanlding MDM command",
				},
			},
		}
	}

	return c.connect(client, nextConnReq)
}
