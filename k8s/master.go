package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	flag "github.com/spf13/pflag"
)

var portnum *int = flag.IntP("port", "p", 7000, "Port # to listen on. Defaults to 7000")
var numNodes *uint64 = flag.Uint64P("nodes", "n", 3, "Number of replicas. Defaults to 3.")

type reqPkt struct {
	inst     string
	ip       string
	retrieve bool
	retC     chan<- resPkt
}

type resPkt struct {
	id uint64
}

type infoPkt struct {
	inst string
	ip   string
	id   uint64
}

type master struct {
	msgC           chan reqPkt
	regNodesByInst map[string]infoPkt
	regNodesById   map[uint64]infoPkt
	identifier     uint64
	resChans       []reqPkt
}

func (m *master) register(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "POST":
		inst := r.PostFormValue("instance")
		ip := r.PostFormValue("ip")

		if len(inst) > 0 && len(ip) > 0 {
			ret := make(chan resPkt, 1)
			m.msgC <- reqPkt{
				inst: inst,
				ip:   ip,
				retC: ret,
			}

			res := <-ret

			fmt.Fprintf(w, "%d", res.id)
		} else {
			http.Error(w, fmt.Sprintf("invalid input: instance=%s and ip=%s", inst, ip), http.StatusBadRequest)
		}

	default:
		http.Error(w, "Only POST supported", http.StatusMethodNotAllowed)
	}
}

func (m *master) nodes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/nodes" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		ret := make(chan resPkt, 1)
		m.msgC <- reqPkt{
			retrieve: true,
			retC:     ret,
		}

		// res := <-ret
		// fmt.Fprintf(w, "%s", res.conf)

	default:
		http.Error(w, "Only GET supported", http.StatusMethodNotAllowed)
	}
}

func (m *master) Run() {
	for {
		pkt := <-m.msgC

		if pkt.retrieve {
			m.resChans = append(m.resChans, pkt)
			if m.identifier == *numNodes {
				log.Printf("returning info for %d nodes...\n", *numNodes)
				m.returnConf()
			} else {
				log.Printf("not enough nodes. queuing for later...\n")
			}
			continue
		}

		if info, exists := m.regNodesByInst[pkt.inst]; exists {
			log.Printf("node already registered %v; received %v\n", info, pkt)
			return
		}

		rID := m.identifier
		m.identifier++

		// log.Printf("container info %v", container)
		nodeInfo := infoPkt{
			inst: pkt.inst,
			ip:   pkt.ip,
			id:   rID,
		}
		log.Printf("registered node %v\n", nodeInfo)
		m.regNodesByInst[nodeInfo.inst] = nodeInfo
		m.regNodesById[nodeInfo.id] = nodeInfo
		m.resChans = append(m.resChans, pkt)

		if m.identifier == *numNodes {
			log.Printf("registered %v nodes...\n", *numNodes)
			m.returnConf()
		} else if m.identifier > *numNodes {
			panic("more nodes registered than expected...")
		}
	}
}

func (m *master) returnConf() {
	// file, err := os.Open("hotstuff.conf")
	// if err != nil {
	// 	log.Panicf("error opening file: %v", err)
	// }

	// scanner := bufio.NewScanner(file)
	// var outBuf strings.Builder
	// idx := uint64(0)
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	if strings.Index(line, "replica") == 0 {
	// 		outBuf.WriteString(
	// 			strings.Replace(line,
	// 				fmt.Sprintf("127.0.0.1:%d;%d", 10000+idx, 20000+idx),
	// 				fmt.Sprintf("%s:7000;8000", m.regNodesById[idx].ip),
	// 				1))
	// 		idx++
	// 	} else {
	// 		outBuf.WriteString(line)
	// 	}
	// 	outBuf.WriteRune('\n')
	// }

	// if err := scanner.Err(); err != nil {
	// 	log.Fatal(err)
	// }
	// file.Close()
	for _, rpkt := range m.resChans {
		rpkt.retC <- resPkt{
			// conf: outBuf.String(),
			id: m.regNodesByInst[rpkt.inst].id,
		}
	}
	m.resChans = nil
}

func main() {
	flag.Parse()

	m := &master{
		msgC: make(chan reqPkt),
		// resChans:       make([]reqPkt, 0, *numNodes),
		regNodesByInst: make(map[string]infoPkt),
		regNodesById:   make(map[uint64]infoPkt),
	}

	go m.Run()

	http.HandleFunc("/register", m.register)
	http.HandleFunc("/nodes", m.nodes)
	http.HandleFunc("/ips", func(w http.ResponseWriter, r *http.Request) {
		var outBuf strings.Builder
		for id, info := range m.regNodesById {
			outBuf.WriteString(fmt.Sprintf("%s\tnode%d\n", info.ip, id))
		}
		fmt.Fprint(w, outBuf.String())
	})

	fs := http.FileServer(http.Dir("mytestnet/"))
	http.Handle("/configfiles/", http.StripPrefix("/configfiles/", fs))

	log.Printf("Waiting for %d pods to join", *numNodes)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portnum), nil))
}
