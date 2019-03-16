package main

import (
	// "bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	// "os"
	//"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/TokTok/go-toxcore-c"
	"github.com/h2non/filetype"
	shell "github.com/ipfs/go-ipfs-api"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

var server = []interface{}{
	"205.185.116.116", uint16(33445), "A179B09749AC826FF01F37A9613F6B57118AE014D4196A0E1105A98F93A54702",
}
var fname = "./toxecho.data"
var debug = true
var nickPrefix = "IpfsEchoUmu"
var statusText = "Umu!"

func get_file_type(hash string) string {
	log.Println("Getting file info for: ", hash)
	ipfs := shell.NewShell("localhost:5001")
	stream, _ := ipfs.Cat(hash)
	// file, _ := ioutil.ReadAll(stream)
	buf := make([]byte, 200)
	io.ReadFull(stream, buf)

	kind, _ := filetype.Match(buf)
	if kind == filetype.Unknown {
		return fmt.Sprintf("Unknown (maybe plain text?) ")
	} else {
		return fmt.Sprintf("%s", kind.MIME.Value)
	}
}

func split_message(r rune) bool {
	return r == '\n' || r == ' '
}

func check_message_for_ipfs(message string) string {
	ipfs := shell.NewShell("localhost:5001")
	log.Println("Ipfs query for file info")
	unixfs, err := ipfs.FileList(message)
	var answer strings.Builder
	if err != nil {
		log.Println("Not unixfs")
	} else {
		answer.WriteString(fmt.Sprintf("(人◕ω◕) Umu! Got additional info for https://ipfs.io/ipfs/%s !\n", unixfs.Hash))
		log.Println("Unixfs: ", unixfs.Hash, unixfs.Size, unixfs.Type, len(unixfs.Links))
		if len(unixfs.Links) > 0 {
			answer.WriteString(fmt.Sprintf("It's a DIRECTORY (ʘ言ʘ╬) Ipfs Gateway Link: https://ipfs.io/ipfs/%s \nTotal files: %d I will show you only some items, and no filetype:\n", unixfs.Hash, len(unixfs.Links)))
			for _, element := range unixfs.Links {
				log.Println(element.Hash, element.Name, element.Size, element.Type)
				if element.Type == "Directory" {
					message := fmt.Sprintf("Ipfs Gateway Link: https://ipfs.io/ipfs/%s \nFile name: /%s/ File type: %s, File size (bytes): %d\n", element.Hash, element.Name, "Directory", element.Size)
					if len(message)+answer.Len() < 1362 {
						answer.WriteString(message)
					}
				} else {
					message := fmt.Sprintf("Ipfs Gateway Link: https://ipfs.io/ipfs/%s \nFile name: /%s/ File type: %s, File size (bytes): %d\n", element.Hash, element.Name, get_file_type(element.Hash), element.Size)
					if len(message)+answer.Len() < 1362 {
						answer.WriteString(message)
					}
				}
			}
		} else {
			answer.WriteString(fmt.Sprintf("File type: %s, File size (bytes): %d\n", get_file_type(unixfs.Hash), unixfs.Size))
		}
		return answer.String()
	}
	answer.WriteString(fmt.Sprintf("Error getting additional info for /%s (┛◉Д◉)┛彡┻━┻", message))
	return answer.String()
}

func answer_for_ipfs(tox_handle *tox.Tox, message_sender interface{}, message string) {
	if !strings.HasPrefix(message, "Qm") {
		log.Println("Not an ipfs hash")
		return
	}

	fast_response := fmt.Sprintf("(人◕ω◕) Umu! Found ipfs link!\n Ipfs Gateway Link: https://ipfs.io/ipfs/%s \n Stand by for additional info, this may take a while (ﾉ◕ヮ◕)ﾉ*:･ﾟ✧", message)
	switch sender := message_sender.(type) {
	case uint32:
		tox_handle.FriendSendMessage(sender, fast_response)
		log.Println("before check_ipfs")
		answer := check_message_for_ipfs(message)
		log.Println("after check_ipfs")
		status, err := tox_handle.FriendSendMessage(sender, answer)
		if err != nil {
			log.Println("Failed to send message to friend: ", status, err)
			tox_handle.FriendSendMessage(sender, fmt.Sprintf("I can't send additional info (┛◉Д◉)┛彡┻━┻ /%s", err))
		}
	case int:
		tox_handle.GroupMessageSend(sender, fast_response)
		answer := check_message_for_ipfs(message)
		status, err := tox_handle.GroupMessageSend(sender, answer)
		if err != nil {
			log.Println("Failed to send message to group: ", status, err)
			tox_handle.GroupMessageSend(sender, fmt.Sprintf("I can't send additional info (┛◉Д◉)┛彡┻━┻ /%s", err))

		}
	default:
		return
	}
}

func send_motd(tox_handle *tox.Tox, groupNumber int) {
	time.Sleep(5 * time.Second)
	status, err := tox_handle.GroupMessageSend(groupNumber, "Hello there! I'm here to watch for your ipfs links ( ◑ω◑☞)☞")
	if err != nil {
		log.Println("Error sending motd to group:", status, err)
	}
}

func main() {
	opt := tox.NewToxOptions()
	if tox.FileExist(fname) {
		data, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Println(err)
		} else {
			opt.Savedata_data = data
			opt.Savedata_type = tox.SAVEDATA_TYPE_TOX_SAVE
		}
	}
	opt.Tcp_port = 33445
	var t *tox.Tox
	for i := 0; i < 5; i++ {
		t = tox.NewTox(opt)
		if t == nil {
			opt.Tcp_port += 1
		} else {
			break
		}
	}

	r, err := t.Bootstrap(server[0].(string), server[1].(uint16), server[2].(string))
	r2, err := t.AddTcpRelay(server[0].(string), server[1].(uint16), server[2].(string))
	if debug {
		log.Println("bootstrap:", r, err, r2)
	}

	pubkey := t.SelfGetPublicKey()
	seckey := t.SelfGetSecretKey()
	toxid := t.SelfGetAddress()
	if debug {
		log.Println("keys:", pubkey, seckey, len(pubkey), len(seckey))
	}
	log.Println("toxid:", toxid)

	friendList := t.SelfGetFriendList()
	log.Println("FriendList:", friendList)
	log.Println()

	defaultName := t.SelfGetName()
	humanName := nickPrefix + toxid[0:5]
	if humanName != defaultName {
		t.SelfSetName(humanName)
	}
	humanName = t.SelfGetName()
	if debug {
		log.Println(humanName, defaultName, err)
	}

	defaultStatusText, err := t.SelfGetStatusMessage()
	if defaultStatusText != statusText {
		t.SelfSetStatusMessage(statusText)
	}
	if debug {
		log.Println(statusText, defaultStatusText, err)
	}

	sz := t.GetSavedataSize()
	sd := t.GetSavedata()
	if debug {
		log.Println("savedata:", sz, t)
		log.Println("savedata", len(sd), t)
	}
	err = t.WriteSavedata(fname)
	if debug {
		log.Println("savedata write:", err)
	}

	// add friend norequest
	fv := t.SelfGetFriendList()
	for _, fno := range fv {
		fid, err := t.FriendGetPublicKey(fno)
		if err != nil {
			log.Println(err)
		} else {
			t.FriendAddNorequest(fid)
		}
	}
	if debug {
		log.Println("add friends:", len(fv))
	}

	// callbacks
	t.CallbackSelfConnectionStatus(func(t *tox.Tox, status int, userData interface{}) {
		if debug {
			log.Println("on self conn status:", status, userData)
		}
	}, nil)
	t.CallbackFriendRequest(func(t *tox.Tox, friendId string, message string, userData interface{}) {
		log.Println(friendId, message)
		num, err := t.FriendAddNorequest(friendId)
		if debug {
			log.Println("on friend request:", num, err)
		}
		if num < 100000 {
			t.WriteSavedata(fname)
		}
	}, nil)
	t.CallbackFriendMessage(func(t *tox.Tox, friendNumber uint32, message string, userData interface{}) {
		if debug {
			log.Println("on friend message:", friendNumber, message)
		}

		message_splitted := []string{}
		if len(message) > 46 {
			message_splitted = strings.FieldsFunc(message, split_message)
		} else {
			message_splitted = []string{message}
		}
		for _, message_element := range message_splitted {
			log.Println("Splitted message part", message_element)
			go answer_for_ipfs(t, friendNumber, message_element)
		}

	}, nil)
	t.CallbackGroupMessage(func(t *tox.Tox, groupNumber int, peerNumber int, message string, userData interface{}) {
		if debug {
			log.Print("Group message:", groupNumber, peerNumber, message)
		}

		message_splitted := []string{}
		if len(message) > 46 {
			message_splitted = strings.FieldsFunc(message, split_message)
		} else {
			message_splitted = []string{message}
		}
		for _, message_element := range message_splitted {
			log.Println("Splitted message part", message_element)
			go answer_for_ipfs(t, groupNumber, message_element)
		}

	}, nil)

	t.CallbackGroupInvite(func(t *tox.Tox, friendNumber uint32, itype uint8, cookie string, userData interface{}) {
		if debug {
			log.Print("Group invite:", friendNumber, itype, cookie)
		}
		deadloko, _ := strconv.ParseInt(<your id here>, 10, 32)
		if friendNumber != uint32(deadloko) {
			t.FriendSendMessage(friendNumber, "You are not allowed to invite me to groups")
		} else {
			t.FriendSendMessage(friendNumber, fmt.Sprintf("Joining group ୧☉□☉୨ coockie: %s", cookie))
			time.Sleep(5 * time.Second)
			groupNumber, err := t.JoinGroupChat(friendNumber, cookie)
			time.Sleep(5 * time.Second)
			if err == nil {
				log.Println("Joined group: ", groupNumber)
				go send_motd(t, groupNumber)
			} else {
				log.Println("Failed joining group: ", err)
			}
		}
	}, nil)
	t.CallbackFriendConnectionStatus(func(t *tox.Tox, friendNumber uint32, status int, userData interface{}) {
		if debug {
			friendId, err := t.FriendGetPublicKey(friendNumber)
			log.Println("on friend connection status:", friendNumber, status, friendId, err)
		}
	}, nil)
	t.CallbackFriendStatus(func(t *tox.Tox, friendNumber uint32, status int, userData interface{}) {
		if debug {
			friendId, err := t.FriendGetPublicKey(friendNumber)
			log.Println("on friend status:", friendNumber, status, friendId, err)
		}
	}, nil)
	t.CallbackFriendStatusMessage(func(t *tox.Tox, friendNumber uint32, statusText string, userData interface{}) {
		if debug {
			friendId, err := t.FriendGetPublicKey(friendNumber)
			log.Println("on friend status text:", friendNumber, statusText, friendId, err)
		}
	}, nil)

	// toxav loops
	go func() {
		shutdown := false
		loopc := 0
		for !shutdown {
			loopc += 1
			time.Sleep(1000 * 50 * time.Microsecond)
		}
	}()

	// toxcore loops
	shutdown := false
	loopc := 0
	itval := 0
	for !shutdown {
		iv := t.IterationInterval()
		if iv != itval {
			if debug {
				if itval-iv > 20 || iv-itval > 20 {
					log.Println("tox itval changed:", itval, iv)
				}
			}
			itval = iv
		}

		t.Iterate()
		status := t.SelfGetConnectionStatus()
		if loopc%5500 == 0 {
			if status == 0 {
				if debug {
					fmt.Print(".")
				}
			} else {
				if debug {
					fmt.Print(status, ",")
				}
			}
		}
		loopc += 1
		time.Sleep(1000 * 50 * time.Microsecond)
	}

	t.Kill()
}

func makekey(no uint32, a0 interface{}, a1 interface{}) string {
	return fmt.Sprintf("%d_%v_%v", no, a0, a1)
}

func _dirty_init() {
	log.Println("ddddddddd")
	tox.KeepPkg()
}
