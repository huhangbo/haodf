package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"log"
	"strconv"
	"sync"
)

var (
	index = "https://www.haodf.com/doctor/list.html?p="
	mutex = sync.RWMutex{}
	wg    = sync.WaitGroup{}
	c     context.Context
)

func getDoctor(url string, doctorIds *[]string) {
	defer wg.Done()
	var (
		nodes []*cdp.Node
	)
	ctx, cancel := chromedp.NewContext(c, chromedp.WithLogf(log.Printf))
	defer cancel()
	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Nodes(".js-fam-doc-li", &nodes, chromedp.ByQueryAll),
	}); err != nil {
		fmt.Println(err)
	}
	for _, item := range nodes {
		tmpUrl := item.Children[0].Children[0].Attributes[1]
		doctorId := tmpUrl[29 : len(tmpUrl)-5]
		mutex.Lock()
		*doctorIds = append(*doctorIds, doctorId)
		mutex.Unlock()
	}
	return
}

func getDoctorDetail(doctorId string) {
	defer wg.Done()
	var (
		nodes  []*cdp.Node
		name   []*cdp.Node
		office []*cdp.Node
		url    = "https://www.haodf.com/doctor/" + doctorId + "/pingjia-zhenliao.html?siftKey=1&p="
	)
	ctx, cancel := chromedp.NewContext(c, chromedp.WithLogf(log.Printf))
	defer cancel()
	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Nodes(".p_num", &nodes, chromedp.ByQueryAll),
		chromedp.Nodes(".doctor-name", &name, chromedp.ByQueryAll),
		chromedp.Nodes(".doctor-faculty a", &office, chromedp.ByQueryAll),
	}); err != nil {
		fmt.Println(err)
	}
	pageString := nodes[len(nodes)-2].Children[0].NodeValue
	total, _ := strconv.ParseInt(pageString, 10, 64)
	nameString := name[0].Children[0].NodeValue
	officeString := office[1].Children[0].NodeValue
	for i := 0; i < int(total); i++ {
		getComment(url+strconv.Itoa(i+1), nameString, doctorId, officeString)
	}
	return
}

func getComment(url, doctorName, doctorId, office string) {
	var nodes []*cdp.Node
	ctx, cancel := chromedp.NewContext(c, chromedp.WithLogf(log.Printf))
	defer cancel()
	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Nodes(".eva-detail", &nodes, chromedp.ByQueryAll),
	}); err != nil {
		fmt.Println(err)
	}
	for _, item := range nodes {
		sql := `INSERT INTO comment (comment, doctor_id, doctor_name, office) VALUES (?, ?, ?, ?)`
		if _, err := db.Exec(sql, item.Children[0].NodeValue, doctorId, doctorName, office); err != nil {
			fmt.Printf("%v", err)
		}
		fmt.Printf("%s\n", item.Children[0].NodeValue)
	}
	return
}

func p() {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)
	c, _ = chromedp.NewExecAllocator(context.Background(), options...)
	var doctorIds []string
	const totalPage = 90
	wg.Add(5)
	for i := 86; i <= totalPage; i++ {
		listUrl := index + strconv.Itoa(i)
		go getDoctor(listUrl, &doctorIds)
	}
	wg.Wait()
	fmt.Printf("%s", doctorIds)
	wg.Add(len(doctorIds))
	for i := 0; i < len(doctorIds); i++ {
		go getDoctorDetail(doctorIds[i])
	}
	wg.Wait()
	return
}

func main() {
	Init()
	defer Close()
	p()
}
