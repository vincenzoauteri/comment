package main

import (
  "strconv"
  "io/ioutil"
  "log"
  "net/http"
  "regexp"
  "fmt"
  "time"
)

var hrefTagRegexp = regexp.MustCompile("link(.*)href=\"")
var imgTagRegexp = regexp.MustCompile("img src=\"")
var url = "https://news.ycombinator.com"


func homeHandler(w http.ResponseWriter, r *http.Request) {
  resp, err := http.Get(url)
  if err != nil {
    log.Fatal(err)
  }
  body, err := ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    log.Fatal(err)
  }

  newBody :=  hrefTagRegexp.ReplaceAll(body, []byte("link" + "${1}" + "href=\"static/"))
  newBody =  imgTagRegexp.ReplaceAll(newBody, []byte("img src=\"https://news.ycombinator.com/"))
  w.Write(newBody)

}

func itemHandler(w http.ResponseWriter, r *http.Request) {

  fmt.Println("Item handler function")
  //Extracts the item id number
  id:= regexp.MustCompile("id=([0-9]*)").FindStringSubmatch(string(r.URL.RawQuery))[1]
  fmt.Println("Item id:"+ id)

  resp, err := http.Get(url+string(r.URL.Path)+"?"+string(r.URL.RawQuery))

  if err != nil {
    log.Fatal(err)
  }

  itemPattern := regexp.MustCompile(":" + id + "=([0-9]*)")

  //Check for cookie with last visit time for items
  cookie, err := r.Cookie("Visits")

  var lastVisited int64 = 0

  if err != nil {
    fmt.Println("Creating new cookie")
    //If cookie does not exists create it
    cookie = new(http.Cookie)
    cookie.Name = "Visits"
  } else {
    //Check if the item id exists in the user's cookie and extracts the unix time
    fmt.Println("Cookie exists")

    found := itemPattern.FindStringSubmatch(cookie.Value)

    if found != nil {
      fmt.Println("last visited:", itemPattern.FindStringSubmatch(cookie.Value))
      lastVisited, err = strconv.ParseInt(itemPattern.FindStringSubmatch(cookie.Value)[1],10,64)
    }
  }


  body, err := ioutil.ReadAll(resp.Body)
  resp.Body.Close()

  if err != nil {
    log.Fatal(err)
  }

  newBody :=  hrefTagRegexp.ReplaceAll(body, []byte("link" + "${1}" + "href=\"static/"))
  newBody =  imgTagRegexp.ReplaceAll(newBody, []byte("img src=\"https://news.ycombinator.com/"))


  timeAgoPattern := regexp.MustCompile("[0-9]+? [A-Za-z]+? ago")
  colorPattern := regexp.MustCompile(" color=#(.+?)>")

  postTimeSlice := timeAgoPattern.FindAllString(string(newBody),-1)[1:]
  postTimeSliceIndex := timeAgoPattern.FindAllStringIndex(string(newBody),-1)[1:]

  for i,v := range postTimeSlice {
    minutePattern:= regexp.MustCompile("([0-9]+?) minute")
    hourPattern:= regexp.MustCompile("([0-9]+?) hour")
    dayPattern:= regexp.MustCompile("([0-9]+?) day")
    submissionTime := time.Now().Unix() 
    if  m := dayPattern.FindStringSubmatch(v); m != nil {
      intDay, _:= strconv.ParseInt(m[1],10,64)
      submissionTime = time.Now().Unix() - 3600*24*intDay
      fmt.Println("comment:",i," submision time:", submissionTime,"last visited:", lastVisited)
    }
    if  m := hourPattern.FindStringSubmatch(v); m != nil {
      intHour, _:= strconv.ParseInt(m[1],10,64)
      submissionTime = time.Now().Unix() - 3600*intHour
      fmt.Println("comment:",i," submision time:", submissionTime,"last visited:", lastVisited )
    }
    if  m := minutePattern.FindStringSubmatch(v); m != nil {
      intMinute, _ := strconv.ParseInt(m[1],10,64)
      submissionTime = time.Now().Unix() - 60*intMinute
      fmt.Println("comment:",i," submision time:", submissionTime,"last visited:", lastVisited )
    }

    if submissionTime >= lastVisited && lastVisited!= 0{
      colorIndexStart := postTimeSliceIndex[i][1];
      colorIndexEnd := len(newBody)
      fmt.Println("i:",i,"len(postTimeSliceIndex[:][0])",len(postTimeSliceIndex) )
      if i+1 < len(postTimeSliceIndex) {
        colorIndexEnd = postTimeSliceIndex[i+1][0];
      }
      newBody = []byte(string(newBody[:colorIndexStart]) + string(colorPattern.ReplaceAll(newBody[colorIndexStart:colorIndexEnd], []byte(" color=#ff0000>"))) +  string(newBody[colorIndexEnd:])) 
    }

  }


  //fmt.Println(postTimeSlice)
  fmt.Println(postTimeSliceIndex)
  //fmt.Println(colorSlice)


  //colorOfNextComment := colorPattern.FindSubmatch(body[index:])[1]

  //fmt.Println( string(colorOfNextComment))



  now := strconv.FormatInt(time.Now().Unix(),10)
  if lastVisited  > 0 {
    fmt.Println("Updating item value")
    //If there was already a time value for the item
    newCookie :=  itemPattern.ReplaceAll([]byte(cookie.Value), []byte(":" + id + "=" + now))
    cookie.Value = string(newCookie)
  } else {
    fmt.Println("creating new item value")
    //If not create a new entry for the ite)m with the current time
    cookie.Value = cookie.Value + ":" + id + "=" + now
  }

  http.SetCookie(w,cookie)
  w.Write(newBody)

}



func main () {
  fmt.Println("Staring Webserver at 8080")
  http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
  http.HandleFunc("/",homeHandler)
  http.HandleFunc("/news",homeHandler)
  http.HandleFunc("/item",itemHandler)
  http.ListenAndServe(":8080",nil)
}
