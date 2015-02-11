package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"regexp"
	"strings"
	//"sync"
	"bytes"
	"net/http"
	"os"
	"time"

	"github.com/Pallinder/go-randomdata"
)

var (
	PORT = os.Getenv("PORT")
)

var (
	regexActLink = regexp.MustCompile(`http://merahputih.com/activation/[a-zA-Z0-9]*`)
)

type DataMail struct {
	EmailAddress string `json:"email_addr"`
	Token        string `json:"sid_token"`
}

type ResponseAct struct {
	List []ActMail `json:"list"`
}

type ResponseActClick struct {
	MailBody string `json:"mail_body"`
}

type ActMail struct {
	MailID   string `json:"mail_id"`
	MailFrom string `json:"mail_from"`
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))

	}
	return string(bytes)

}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)

}

func Activate(buffer *bytes.Buffer) (string, error) {
	//defer wg.Done()
	genders := []string{"m", "f"}
	// create random data
	firstname := randomdata.FirstName(randomdata.RandomGender)
	lastname := randomdata.LastName()
	//city := randomdata.City()
	gender := genders[rand.Intn(1)]
	username := randomString(12)
	password := "" + firstname + "" + lastname
	//fmt.Println(username)

	/*
		fmt.Println("Firstname: ", firstname)
		fmt.Println("Lastname: ", lastname)
		fmt.Println("username: ", lastname+city)
		fmt.Println("gender: ", genders[rand.Intn(1)])
	*/

	// get data of email
	resp, err := http.Get("http://api.guerrillamail.com/ajax.php?f=set_email_user&email_user=" + lastname + firstname + randomdata.LastName() + "&domain=guerrillamail.com&lang=en&ip=127.0.0.1&agent=Mozilla_foo_bar")
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	// fmt.Printf("JSON: %s \n", body)
	datamail := &DataMail{}
	json.Unmarshal(body, &datamail)
	// fmt.Println(datamail.EmailAddress, datamail.Token)

	// register a user
	client := http.Client{}
	req, err := http.NewRequest("POST", "http://merahputih.com/do-register", strings.NewReader("firstname="+firstname+"&lastname="+lastname+"&gender="+gender+"&username="+username+"&email="+datamail.EmailAddress+"&password="+password+"&repassword="+password))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8,id;q=0.6,ms;q=0.4")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Host", "merahputih.com")
	req.Header.Add("Origin", "http://merahputih.com")
	req.Header.Add("Referer", "http://merahputih.com/register")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux i686) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.93 Safari/537.36")
	_, err = client.Do(req)
	if err != nil {
		return "", err
	}
	// fmt.Println(respReg.Status)

	// check email
	time.Sleep(10 * time.Second)
	respAct, err := http.Get("http://api.guerrillamail.com/ajax.php?f=get_email_list&ip=127.0.0.1&agent=Mozilla_foo_bar&offset=0&sid_token=" + datamail.Token)
	if err != nil {
		return "", err
	}
	bodyAct, err := ioutil.ReadAll(respAct.Body)
	// fmt.Printf("JSON: %s \n", bodyAct)
	responseact := &ResponseAct{}
	json.Unmarshal(bodyAct, &responseact)
	// fmt.Printf("%+v\n", responseact)

	// get link confirmation
	emailid := responseact.List[0].MailID
	respActClick, err := http.Get("http://api.guerrillamail.com/ajax.php?f=fetch_email&sid_token=" + datamail.Token + "&email_id=" + emailid)
	if err != nil {
		return "", err
	}
	bodyActClick, err := ioutil.ReadAll(respActClick.Body)
	// fmt.Printf("JSON: %s \n", bodyActClick)
	responseactclick := &ResponseActClick{}
	json.Unmarshal(bodyActClick, &responseactclick)
	linkActivation := regexActLink.FindString(responseactclick.MailBody)
	// fmt.Println(linkActivation)
	_, err = http.Get(linkActivation)
	if err != nil {
		return "", err
	}
	// log.Println(activate.Status)
	//fmt.Printf("%s:%s\n", username, password)
	result := "" + username + ":" + password + "\n"
	_, err = buffer.WriteString(result)
	if err != nil {
		return "", err
	}
	return result, nil
	//done <- 1
}

func main() {
	//done := make(chan int)
	//var wg sync.WaitGroup
	var buffer bytes.Buffer

	//wg.Wait()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", buffer.String())
	})

	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		result, err := Activate(&buffer)
		if err != nil {
			fmt.Fprintf(w, "%v", err)
		}
		fmt.Fprintf(w, "%s", result)
	})
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}
