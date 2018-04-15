package sip

import (
    "net"
    "testing"
    "time"
    "fmt"

    "github.com/stretchr/testify/assert"

    "github.com/elastic/beats/libbeat/beat"
    "github.com/elastic/beats/libbeat/common"
    "github.com/elastic/beats/libbeat/logp"
    "github.com/elastic/beats/packetbeat/protos"
)


func TestInit(t *testing.T) {
    // TODO: Is it need test implementation?
}

func TestSetFromConfig(t *testing.T) {
    sip:=sipPlugin{}
    cfg:=sipConfig{}
    cfg.Ports=[]int{5060,5061}
    cfg.ParseDetail=true

    sip.setFromConfig(&cfg)
    assert.Equal(t, 5060 , sip.ports[0]      , "There should be included 5060." )
    assert.Equal(t, 5061 , sip.ports[1]      , "There should be included 5061." )
    assert.Equal(t, 2     , len(sip.ports)   , "There should be included 5060." )
    assert.Equal(t, true , sip.parseDetail   , "There should be included 5061." )
}

func TestSetFromConfigDefault(t *testing.T) {
    sip:=sipPlugin{}
    cfg:=sipConfig{}

    sip.setFromConfig(&cfg)
    //assert.Equal(t, 5060  , sip.ports[0]     , "There should be included 5060." )
    assert.Equal(t, 0     , len(sip.ports)   , "There should be included 5060." )
    assert.Equal(t, false , sip.parseDetail  , "There should be included 5061." )
}

func TestGetPorts(t *testing.T) {
    sip:=sipPlugin{}
    sip.ports=[]int{5060,5061,1123,5555}
    ports:=sip.GetPorts()

    assert.Equal(t, 5060 , ports[0]          , "There should be included 5060." )
    assert.Equal(t, 5061 , ports[1]          , "There should be included 5061." )
    assert.Equal(t, 1123 , ports[2]          , "There should be included 5061." )
    assert.Equal(t, 5555 , ports[3]          , "There should be included 5061." )
}

func TestPublishMessage(t *testing.T) {
    sip:=sipPlugin{}

    raw_text   :="test raw string"
    method_text:="INVITE"
    phrase_text:="OK"
    ipTuple := common.NewIPPortTuple(4,
        net.ParseIP("10.0.0.1"), 1111,
        net.ParseIP("10.0.0.2"), 2222)
    msg:=sipMessage{transport:0, raw:common.NetString(raw_text),
                    tuple: ipTuple , method: common.NetString(method_text),
                    requestUri: common.NetString("sip:test"),
                    statusCode: uint16(200), statusPhrase: common.NetString(phrase_text),
                    from: common.NetString("from"), to: common.NetString("to"),
                    cseq: common.NetString("cseq"), callid: common.NetString("callid"),
                    contentlength: 10}

    // avoid to sip.results initialization error
    sip.publishMessage(&msg)
    assert.Nil(t, sip.results , "sip.results should still nil." )

    store:= &eventStore{}

    sip.results=store.publish
    sip.publishMessage(&msg)
    assert.Equal(t, 1, store.size() , "There should be added one packet in store after publish." )
    assert.Equal(t, phrase_text, store.events[0].Fields["sip.status-phrase"], "Compare published packet and stored data." )
    assert.Equal(t, nil        , store.events[0].Fields["sip.method"]       , "Compare published packet and stored data." )
    assert.Equal(t, raw_text   , store.events[0].Fields["sip.raw"]          , "Compare published packet and stored data." )
}
func TestPublishMessageWithDetailOption(t *testing.T) {
    sip:=sipPlugin{}
    sip.parseDetail=true
    raw_text   :="test raw string"
    method_text:="INVITE"
    from:=`"0311112222"<sip:311112222@sip.addr:5060>;tag=FromTag`
    to:=`<sip:612341234@192.168.0.1>`
    request_uri:="sip:+8137890123;npdi;rn=+81312341234@hoge.com:5060;user=phone;transport=udp"
    cseq_num:=6789
    cseq_method:="INVITE"
    cseq:=fmt.Sprintf("%d %s",cseq_num,cseq_method)
    ipTuple := common.NewIPPortTuple(4,
        net.ParseIP("10.0.0.1"), 1111,
        net.ParseIP("10.0.0.2"), 2222)

    headers := map[string][]common.NetString{}
    to_h:=[]common.NetString{}
    to_h=append(to_h,common.NetString(to))
    from_h:=[]common.NetString{}
    from_h=append(from_h,common.NetString(from))
    pai_h:=[]common.NetString{}
    pai_0:=`"0312341234" <tel:+81312341234;cpc=ordinary>`
    pai_h=append(pai_h,common.NetString(`"0312341234" <tel:+81312341234;cpc=ordinary>`))
    pai_1:=`<sip:+81312341234@hoge.com;user=phone;cpc=ordinary>`
    pai_h=append(pai_h,common.NetString(`<sip:+81312341234@hoge.com;user=phone;cpc=ordinary>`))
    callid:=`1-2363@192.168.122.252`
    callid_h:=[]common.NetString{}
    callid_h=append(callid_h,common.NetString(callid))
    headers["from"]=from_h
    headers["call-id"]=callid_h
    headers["p-asserted-identity"]=pai_h
    headers["to"]=to_h
    headers["orig"]=to_h

    msg:=sipMessage{transport:0,     raw:common.NetString(raw_text),
                    tuple: ipTuple , method: common.NetString(method_text),
                    requestUri: common.NetString(request_uri),
                    from: common.NetString(from),
                    to: common.NetString(to),
                    cseq: common.NetString(cseq),
                    callid: common.NetString("callid"),
                    contentlength: 10,isRequest:true}
    msg.headers= &headers

    store:= &eventStore{}
    sip.results=store.publish
    sip.publishMessage(&msg)

    stored:=store.events[0].Fields
    assert.Equal(t, method_text, stored["sip.method"     ], "Invalid Method text" )
    assert.Equal(t, request_uri, stored["sip.request-uri"], "Invalid Request URI" )
    assert.Equal(t, to         , stored["sip.to"         ], "Invalid To text"     )
    assert.Equal(t, from       , stored["sip.from"       ], "Invalid from text"   )
    assert.Equal(t, cseq       , stored["sip.cseq"       ], "Invalid CSeq text"   )
    assert.Equal(t, request_uri, stored["sip.request-uri"], "Invalid Request URI" )

    userpart:="+8137890123;npdi;rn=+81312341234"
    assert.Equal(t, userpart  ,     stored["sip.request-uri.user"]                       , "Invalid Request URI user info"        )
    assert.Equal(t, "hoge.com",     stored["sip.request-uri.host"]                       , "Invalid Request URI host"             )
    assert.Equal(t, 5060      ,     stored["sip.request-uri.port"]                       , "Invalid Request URI host"             )
    assert.Contains(t         ,     stored["sip.request-uri.uri-params"], "user=phone"   , "Invalid Request URI parameter"        )
    assert.Contains(t         ,     stored["sip.request-uri.uri-params"], "transport=udp", "Invalid Request URI parameter"        )
    assert.Equal(t, 2         , len(stored["sip.request-uri.uri-params"].([]string))     , "Invalid Request URI parameter length" )

    assert.Equal(t, "0311112222",     stored["sip.from.display"   ]               , "Invalid from display"    )
    assert.Equal(t, "311112222" ,     stored["sip.from.user"      ]               , "Invalid from user"       )
    assert.Equal(t, "sip.addr"  ,     stored["sip.from.host"      ]               , "Invalid from host"       )
    assert.Equal(t, 5060        ,     stored["sip.from.port"      ]               , "Invalid from port"       )
    assert.Equal(t, 1           , len(stored["sip.from.params"    ].([]string))   , "Invalid from params"     )
    assert.Contains(t,                stored["sip.from.params"    ], "tag=FromTag", "Invalid from params"     )
    assert.Equal(t, nil         ,     stored["sip.from.uri-params"]               , "Invalid from uri-params" )

    headers_p:=(stored["sip.headers"].(common.MapStr))["from"].([]common.MapStr)
    assert.Equal(t, common.NetString(from),     headers_p[0]["raw"       ]              , "Invalid from text"       )
    assert.Equal(t, "0311112222"          ,     headers_p[0]["display"   ]              , "Invalid from display"    )
    assert.Equal(t, "311112222"           ,     headers_p[0]["user"      ]              , "Invalid from user"       )
    assert.Equal(t, "sip.addr"            ,     headers_p[0]["host"      ]              , "Invalid from host"       )
    assert.Equal(t, 5060                  ,     headers_p[0]["port"      ]              , "Invalid from port"       )
    assert.Contains(t                     ,     headers_p[0]["params"    ],"tag=FromTag", "Invalid from params"     )
    assert.Equal(t, 1                     , len(headers_p[0]["params"    ].([]string))  , "Invalid from params"     )
    assert.Equal(t, nil                   ,     headers_p[0]["uri-params"]              , "Invalid from uri-params" )

    assert.Equal(t, nil          , stored["sip.to.display"     ], "Invalid to display"    )
    assert.Equal(t, "612341234"  , stored["sip.to.user"        ], "Invalid to user"       )
    assert.Equal(t, "192.168.0.1", stored["sip.to.host"        ], "Invalid to host"       )
    assert.Equal(t, nil          , stored["sip.to.port"        ], "Invalid to port"       )
    assert.Equal(t, nil          , stored["sip.to.params"      ], "Invalid to params"     )
    assert.Equal(t, nil          , stored["sip.to.uri-params"  ], "Invalid to uri-params" )

    headers_p =(stored["sip.headers"].(common.MapStr))["to"].([]common.MapStr)
    assert.Equal(t, common.NetString(to)  , headers_p[0]["raw"        ],"Invalid to text"       )
    assert.Equal(t, nil                   , headers_p[0]["display"    ],"Invalid to display"    )
    assert.Equal(t, "612341234"           , headers_p[0]["user"       ],"Invalid to user"       )
    assert.Equal(t, "192.168.0.1"         , headers_p[0]["host"       ],"Invalid to host"       )
    assert.Equal(t, nil                   , headers_p[0]["port"       ],"Invalid to port"       )
    assert.Equal(t, nil                   , headers_p[0]["params"     ],"Invalid to params"     )
    assert.Equal(t, nil                   , headers_p[0]["uri-params" ],"Invalid to uri-params" )

    headers_p =(stored["sip.headers"].(common.MapStr))[  "p-asserted-identity"].([]common.MapStr)
    assert.Equal(t, common.NetString(pai_0), headers_p[0]["raw"        ],"Invalid p-asserted-identity text"       )
    assert.Equal(t, "0312341234"           , headers_p[0]["display"    ],"Invalid p-asserted-identity display"    )
    assert.Equal(t, nil                    , headers_p[0]["user"       ],"Invalid p-asserted-identity user"       )
    assert.Equal(t, "+81312341234"         , headers_p[0]["host"       ],"Invalid p-asserted-identity host"       )
    assert.Equal(t, nil                    , headers_p[0]["port"       ],"Invalid p-asserted-identity port"       )
    assert.Equal(t, nil                    , headers_p[0]["params"     ],"Invalid p-asserted-identity params"     )
    assert.Contains(t,     headers_p[0]["uri-params"], "cpc=ordinary"   ,"Invalid p-asserted-identity uri-params" )
    assert.Equal(t, 1, len(headers_p[0]["uri-params"].([]string))       ,"Invalid p-asserted-identity uri-params" )
    assert.Equal(t, common.NetString(pai_1), headers_p[1]["raw"        ],"Invalid p-asserted-identity text"       )
    assert.Equal(t, nil                    , headers_p[1]["display"    ],"Invalid p-asserted-identity display"    )
    assert.Equal(t, "+81312341234"         , headers_p[1]["user"       ],"Invalid p-asserted-identity user"       )
    assert.Equal(t, "hoge.com"             , headers_p[1]["host"       ],"Invalid p-asserted-identity host"       )
    assert.Equal(t, nil                    , headers_p[1]["port"       ],"Invalid p-asserted-identity port"       )
    assert.Equal(t, nil                    , headers_p[1]["params"     ],"Invalid p-asserted-identity params"     )
    assert.Contains(t,     headers_p[1]["uri-params"],"cpc=ordinary"    ,"Invalid p-asserted-identity uri-params" )
    assert.Contains(t,     headers_p[1]["uri-params"],"user=phone"      ,"Invalid p-asserted-identity uri-params" )
    assert.Equal(t, 2, len(headers_p[1]["uri-params"].([]string))       ,"Invalid p-asserted-identity uri-params" )

    headers_p =(stored["sip.headers"].(common.MapStr))[   "call-id"].([]common.MapStr)
    assert.Equal(t, common.NetString(callid), headers_p[0]["raw"        ],"Invalid call-id text"       )
    assert.Equal(t, nil                     , headers_p[0]["display"    ],"Invalid call-id display"    )
    assert.Equal(t, nil                     , headers_p[0]["user"       ],"Invalid call-id user"       )
    assert.Equal(t, nil                     , headers_p[0]["host"       ],"Invalid call-id host"       )
    assert.Equal(t, nil                     , headers_p[0]["port"       ],"Invalid call-id port"       )
    assert.Equal(t, nil                     , headers_p[0]["params"     ],"Invalid call-id params"     )
    assert.Equal(t, nil                     , headers_p[0]["uri-params" ],"Invalid call-id uri-params" )

    headers_p =(stored["sip.headers"].(common.MapStr))["orig"].([]common.MapStr)
    assert.Equal(t, common.NetString(to) , headers_p[0]["raw"        ],"Invalid orig text"       )
    assert.Equal(t, nil                  , headers_p[0]["display"    ],"Invalid orig display"    )
    assert.Equal(t, nil                  , headers_p[0]["user"       ],"Invalid orig user"       )
    assert.Equal(t, nil                  , headers_p[0]["host"       ],"Invalid orig host"       )
    assert.Equal(t, nil                  , headers_p[0]["port"       ],"Invalid orig port"       )
    assert.Equal(t, nil                  , headers_p[0]["params"     ],"Invalid orig params"     )
    assert.Equal(t, nil                  , headers_p[0]["uri-params" ],"Invalid orig uri-params" )

    assert.Equal(t, cseq_num     , stored["sip.cseq.number"], "Invalid cseq number" )
    assert.Equal(t, cseq_method  , stored["sip.cseq.method"], "Invalid cseq method" )
}

func TestCreateSIPMessage(t *testing.T) {
    sip:=sipPlugin{}
    var trans transport
    trans=transportTCP
    garbage := []byte( "Go is an open source programming language "   +
                       "that makes it easy to build simple, reliable, "+
                       "and efficient software.")
    sipMsg,err:=sip.createSIPMessage(trans,garbage)

    assert.Nil(t, err, "Should be no errors." )
    assert.Equal(t, trans, sipMsg.transport, "Compare transport value." )
    assert.Equal(t, garbage, sipMsg.raw, "Compare packet raw message." )
    assert.Equal(t, -1, sipMsg.hdr_start    , "Initialization check." )
    assert.Equal(t, -1, sipMsg.hdr_len      , "Initialization check." )
    assert.Equal(t, -1, sipMsg.bdy_start    , "Initialization check." )
    assert.Equal(t, -1, sipMsg.contentlength, "Initialization check." )
}

// Test Cases migrated from sip_test.go 2018-03-03
// Test Constants
const (
    serverIP   = "192.168.0.1"
    serverPort = 5060
    clientIP   = "10.0.0.1"
    clientPort = 5060
)

// Request and response addresses.
var (
    forward = common.NewIPPortTuple(4,
        net.ParseIP(serverIP), serverPort,
        net.ParseIP(clientIP), clientPort)
    reverse = common.NewIPPortTuple(4,
        net.ParseIP(clientIP), clientPort,
        net.ParseIP(serverIP), serverPort)
)

type eventStore struct {
    events []beat.Event
}

func (e *eventStore) publish(event beat.Event) {
    e.events = append(e.events, event)
}

func (e *eventStore) empty() bool {
    return len(e.events) == 0
}

func (e *eventStore) size() int {
    return len(e.events)
}

func newSIP(store *eventStore, verbose bool) *sipPlugin {
    level := logp.WarnLevel
    if verbose {
        level = logp.DebugLevel
    }
    logp.DevelopmentSetup(
        logp.WithLevel(level),
        logp.WithSelectors("sip"),
    )

    callback := func(beat.Event) {}
    if store != nil {
        callback = store.publish
    }

    cfg, _ := common.NewConfigFrom(map[string]interface{}{
        "ports":               []int{serverPort},
    })
    sip, err := New(false, callback, cfg)
    if err != nil {
        panic(err)
    }

    return sip.(*sipPlugin)
}

func newPacket(t common.IPPortTuple, payload []byte) *protos.Packet {
    return &protos.Packet{
        Ts:      time.Now(),
        Tuple:   t,
        Payload: payload,
    }
}

// Verify that an empty packet is safely handled (no panics).
func TestParseUdp_emptyPacket(t *testing.T) {
    store := &eventStore{}
    sip := newSIP(store, testing.Verbose())
    packet := newPacket(forward, []byte{})
    sip.ParseUDP(packet)

    assert.Equal(t, 0, store.size(), "There should be one message published.")
}

// Verify that a malformed packet is safely handled (no panics).
func TestParseUdp_malformedPacket(t *testing.T) {
    store := &eventStore{}
    sip := newSIP(store, testing.Verbose())
    garbage := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
    packet := newPacket(forward, garbage)
    sip.ParseUDP(packet)

    assert.Equal(t, 0, store.size(), "There should be one message published.")
}

func TestParseUdp_requestPacketWithSDP(t *testing.T){
    store := &eventStore{}
    sip := newSIP(store, testing.Verbose())
    garbage := []byte( "INVITE sip:0312345678@192.168.0.1;user=phone SIP/2.0\r\n"   +
                       "Via: SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK81075720\r\n"  +
                       "From: <sip:sipurl@192.168.0.1>;tag=269050131\r\n"           +
                       "To: <sip:0312341234@192.168.0.1;user=phone>\r\n"            +
                       "Contact: <sip:301234123@10.0.0.1;user=phone>\r\n"           +
                       "Call-ID: hogehoge@192.168.0.1\r\n"                          +
                       "CSeq: 1 INVITE\r\n"                                         +
                       "Max-Forwards: 70\r\n"                                       +
                       "Allow: INVITE, ACK, CANCEL, BYE, UPDATE, PRACK\r\n"         +
                       "Supported: 100rel,timer\r\n"                                +
                       "Session-Expires: 300\r\n"                                   +
                       "Privacy: none\r\n"                                          +
                       "P-Preferred-Identity: <tel:0387654321>\r\n"                 +
                       "Content-Type: application/sdp\r\n"                          +
                       "Content-Length: 107\r\n"                                    +
                       "\r\n"                                                       +
                       "v=0\r\n"                                                    +
                       "o=- 0 0 IN IP4 10.0.0.1\r\n"                                +
                       "s=-\r\n"                                                    +
                       "c=IN IP4 10.0.0.1\r\n"                                      +
                       "t=0 0\r\n"                                                  +
                       "m=audio 5012 RTP/AVP 0\r\n"                                 +
                       "a=rtpmap:0 PCMU/8000\r\n")
    packet := newPacket(forward, garbage)
    sip.ParseUDP(packet)
    assert.Equal(t, 1, store.size(), "There should be one message published.")
    if store.size() == 1{
        fields:=store.events[0].Fields
        headers,_:=fields["sip.headers"].(common.MapStr)
        // mandatories
        assert.Equal(t, "INVITE",
                        fields["sip.method"], 
                        "There should be [INVITE].")

                        assert.Equal(t, "sip:0312345678@192.168.0.1;user=phone", 
                        fields["sip.request-uri"], 
                        "There should be [sip:0312345678@192.168.0.1;user=phone].")

        assert.Equal(t, "hogehoge@192.168.0.1" ,
                        fields["sip.call-id"], 
                        "There should be [hogehoge@192.168.0.1].")

        assert.Equal(t, "<sip:sipurl@192.168.0.1>;tag=269050131",
                        fields["sip.from"], 
                        "There should be [<sip:sipurl@192.168.0.1>;tag=269050131].")

        assert.Equal(t, "<sip:0312341234@192.168.0.1;user=phone>",
                        fields["sip.to"],
                        "There should be [<sip:0312341234@192.168.0.1;user=phone>].")

        assert.Equal(t, "1 INVITE",
                        fields["sip.cseq"],
                        "There should be [1 INVITE].")
        // headers
        assert.Equal(t, "application/sdp",
                        fmt.Sprintf("%s",(headers["content-type"].([]common.NetString))[0]),
                        "There should be [application/sdp].")

        assert.Equal(t, "70",
                        fmt.Sprintf("%s",(headers["max-forwards"].([]common.NetString))[0]),
                        "There should be [70].")

        assert.Contains(t, headers["allow"], common.NetString("INVITE"), "There should contain Allow headers.")
        assert.Contains(t, headers["allow"], common.NetString("ACK")   , "There should contain Allow headers.")
        assert.Contains(t, headers["allow"], common.NetString("CANCEL"), "There should contain Allow headers.")
        assert.Contains(t, headers["allow"], common.NetString("BYE")   , "There should contain Allow headers.")
        assert.Contains(t, headers["allow"], common.NetString("UPDATE"), "There should contain Allow headers.")
        assert.Contains(t, headers["allow"], common.NetString("PRACK") , "There should contain Allow headers.")

        assert.Contains(t, headers["supported"], common.NetString("100rel"), "There should contain Supported headers.")
        assert.Contains(t, headers["supported"], common.NetString("timer") , "There should contain Supported headers.")

        assert.Equal(t, "300",
                        fmt.Sprintf("%s",(headers["session-expires"].([]common.NetString))[0]),
                        "There should be [300].")

        assert.Equal(t, "none",
                        fmt.Sprintf("%s",(headers["privacy"].([]common.NetString))[0]),
                        "There should be [none].")

        assert.Equal(t, "<tel:0387654321>",
                        fmt.Sprintf("%s",(headers["p-preferred-identity"].([]common.NetString))[0]),
                        "There should be [<tel:0387654321>].")

        assert.Equal(t, "107",
                        fmt.Sprintf("%s",(headers["content-length"].([]common.NetString))[0]),
                        "There should be [107].")


        via0:="SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK81075720"
        assert.Equal(t, via0,
                        fmt.Sprintf("%s",(headers["via"].([]common.NetString))[0]),
                        fmt.Sprintf("There should be [%s].",via0))
    }
}

func TestParseUdp_requestPacketWithoutSDP(t *testing.T){
    store := &eventStore{}
    sip := newSIP(store, testing.Verbose())
    garbage := []byte(  "ACK sip:0312345678@192.168.0.1:5060 SIP/2.0\r\n"               +
                        "Via: SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK3408987398\r\n"   +
                        "From: <sip:hogehoge@example.com>;tag=5408647717\r\n"           +
                        "To: <sip:0312345678@192.168.0.1>;tag=3713480994\r\n"           +
                        "Call-ID: hogehoge@10.0.0.1\r\n"                                +
                        "CSeq: 1 ACK\r\n"                                               +
                        "Content-Length: 0\r\n"                                         +
                        "Max-Forwards: 70\r\n"                                          +
                        "\r\n")
    packet := newPacket(forward, garbage)
    sip.ParseUDP(packet)
    assert.Equal(t, 1, store.size(), "There should be one message published.")
}

func TestParseUdp_requestPacketBeforeStartCRLF(t *testing.T){
    store := &eventStore{}
    sip := newSIP(store, testing.Verbose())
    garbage := []byte(  "\r\n"                                                          +
                        "\r\n"                                                          +
                        "\r\n"                                                          +
                        "\r\n"                                                          +
                        "\r\n"                                                          +
                        "ACK sip:0312345678@192.168.0.1:5060 SIP/2.0\r\n"               +
                        "Via: SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK3408987398\r\n"   +
                        "From: <sip:hogehoge@example.com>;tag=5408647717\r\n"           +
                        "To: <sip:0312345678@192.168.0.1>;tag=3713480994\r\n"           +
                        "Call-ID: hogehoge@10.0.0.1\r\n"                                +
                        "CSeq: 1 ACK\r\n"                                               +
                        "Content-Length: 0\r\n"                                         +
                        "Max-Forwards: 70\r\n"                                          +
                        "\r\n")
    packet := newPacket(forward, garbage)
    sip.ParseUDP(packet)
    assert.Equal(t, 1, store.size(), "There should be one message published.")
}


func TestParseUdp_responsePacketWithSDP(t *testing.T){
    store := &eventStore{}
    assert.Equal(t, 0, store.size(), "There should be one message published.")
    sip := newSIP(store, testing.Verbose())
    garbage := []byte( "SIP/2.0 183 Session Progess\r\n"                                                      +
                       "Via: SIP/2.0/UDP cw-aio:5060;rport;branch=z9hG4bKPjsRBrmG2vdijbHibFAGTin3eIn6pWysl1," +
                       " SIP/2.0/TCP 192.168.0.1:5052;rport=55196;received=192.168.0.1;"                      +
                       "branch=z9hG4bKPjzjRVAiigVbR6RMhBFOgNh6BXHP80-aBf,"                                    +
                       " SIP/2.0/TCP 192.168.0.1:5058;rport=34867;received=192.168.0.1;"                      +
                       "branch=z9hG4bKPjkp40B7iQTntn1rf9TuASHKtyhPss8fh5,"                                    +
                       " SIP/2.0/UDP 10.0.0.1:5060;received=10.0.0.1;branch=z9hG4bK-2363-1-0\r\n"             +
                       "From: \"sipp\" <sip:sipp@10.0.0.1>;tag=2363SIPpTag001\r\n"                            +
                       "To: \"sut\" <sip:6505550252@192.168.0.1>;tag=16489SIPpTag012\r\n"                     +
                       "Call-ID: 1-2363@10.0.0.1\r\n"                                                         +
                       "CSeq: 1 INVITE\r\n"                                                                   +
                       "Contact: <sip:192.168.0.1:5060;transport=UDP>\r\n"                                    +
                       "Content-Type: application/sdp\r\n"                                                    +
                       "Content-Length: 114\r\n"                                                              +
                       "\r\n"                                                                                 +
                       "v=0\r\n"                                                                              +
                       "o=- 0 0 IN IP4 192.168.0.1\r\n"                                                       +
                       "s=-\r\n"                                                                              +
                       "c=IN IP4 192.168.0.1\r\n"                                                             +
                       "t=0 0\r\n"                                                                            +
                       "m=audio 65000 RTP/AVP 0\r\n"                                                          +
                       "a=rtpmap:0 PCMU/8000\r\n")

    packet := newPacket(forward, garbage)
    sip.ParseUDP(packet)
    assert.Equal(t, 1, store.size(), "There should be one message published.")
    if store.size() == 1{
        fields:=store.events[0].Fields
        headers,_:=fields["sip.headers"].(common.MapStr)
        // mandatories
        assert.Equal(t, "Session Progess",
                        fields["sip.status-phrase"], 
                        "There should be [Session Progress].")

        assert.Equal(t, 183, 
                        fields["sip.status-code"], 
                        "There should be 183.")

        assert.Equal(t, "1-2363@10.0.0.1" ,
                        fields["sip.call-id"], 
                        "There should be [1-2363@10.0.0.1].")

        assert.Equal(t, "\"sipp\" <sip:sipp@10.0.0.1>;tag=2363SIPpTag001",
                        fields["sip.from"], 
                        "There should be [\"sipp\" <sip:sipp@10.0.0.1>;tag=2363SIPpTag001].")

        assert.Equal(t, "\"sut\" <sip:6505550252@192.168.0.1>;tag=16489SIPpTag012",
                        fields["sip.to"],
                        "There should be [\"sut\" <sip:6505550252@192.168.0.1>;tag=16489SIPpTag012].")

        assert.Equal(t, "1 INVITE",
                        fields["sip.cseq"],
                        "There should be [1 INVITE].")
        // headers
        assert.Equal(t, "application/sdp",
                        fmt.Sprintf("%s",(headers["content-type"].([]common.NetString))[0]),
                        "There should be [application/sdp].")

        via0:="SIP/2.0/UDP cw-aio:5060;rport;branch=z9hG4bKPjsRBrmG2vdijbHibFAGTin3eIn6pWysl1"
        assert.Equal(t, via0,
                        fmt.Sprintf("%s",(headers["via"].([]common.NetString))[0]),
                        fmt.Sprintf("There should be [%s].",via0))

        via1:="SIP/2.0/TCP 192.168.0.1:5052;rport=55196;received=192.168.0.1;" +
              "branch=z9hG4bKPjzjRVAiigVbR6RMhBFOgNh6BXHP80-aBf"
        assert.Equal(t, via1,
                        fmt.Sprintf("%s",(headers["via"].([]common.NetString))[1]),
                        fmt.Sprintf("There should be [%s].",via1))
via2:="SIP/2.0/TCP 192.168.0.1:5058;rport=34867;received=192.168.0.1;" +
              "branch=z9hG4bKPjkp40B7iQTntn1rf9TuASHKtyhPss8fh5"
        assert.Equal(t, via2,
                        fmt.Sprintf("%s",(headers["via"].([]common.NetString))[2]),
                        fmt.Sprintf("There should be [%s].",via2))

        via3:="SIP/2.0/UDP 10.0.0.1:5060;received=10.0.0.1;branch=z9hG4bK-2363-1-0"
        assert.Equal(t, via3,
                        fmt.Sprintf("%s",(headers["via"].([]common.NetString))[3]),
                        fmt.Sprintf("There should be [%s].",via3))
    }
}

func TestParseUdp_responsePacketWithoutSDP(t *testing.T){
    store := &eventStore{}
    sip := newSIP(store, testing.Verbose())
    garbage := []byte( "SIP/2.0 407 Proxy Authentication Required\r\n"              +
                       "Via: SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK3408987398\r\n"+
                       "From: <sip:hogehoge@10.0.0.1>;tag=5408647717\r\n"           +
                       "To: <sip:0312345678@192.168.0.1>;tag=3713480994\r\n"        +
                       "Call-ID: hogehoge@192.168.0.1\r\n"                          +
                       "CSeq: 1 INVITE\r\n"                                         +
                       "Content-Length: 0\r\n"                                      +
                       "Date: Mon, 04 Sep 2017 02:29:54 GMT\r\n"                    +
                       "Proxy-Authenticate: Digest realm=\"example.com\","          +
                       " domain=\"sip:192.168.0.1\", nonce=\"15044921123142536\","  +
                       " opaque=\"\", stale=FALSE, algorithm=MD5\r\n"               +
                       "\r\n")
    packet := newPacket(forward, garbage)
    sip.ParseUDP(packet)

    assert.Equal(t, 1, store.size(), "There should be one message published.")
}

func TestParseUdp_IncompletePacketInBody(t *testing.T) {
    store := &eventStore{}
    sip := newSIP(store, testing.Verbose())
    garbage := []byte(  "INVITE sip:0312345678@192.168.0.1:5060 SIP/2.0\r\n"                        +
                         "Via: SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK1701109339\r\n"               +
                         "From: <sip:hogehoge@example.cm>;tag=1451088881\r\n"                        +
                         "To: <sip:0312345678@192.168.0.1>\r\n"                                      +
                         "Call-ID: hogehoge@10.0.0.1\r\n"                                            +
                         "CSeq: 2 INVITE\r\n"                                                        +
                         "Contact: <sip:1833176976@10.0.0.1:5060;transport=udp>\r\n"                 +
                         "Supported: 100rel, timer\r\n"                                              +
                         "Allow: INVITE, ACK, CANCEL, BYE, UPDATE, PRACK\r\n"                        +
                         "Content-Length: 134\r\n"                                                   +
                         "Session-Expires: 180\r\n"                                                  +
                         "Via: SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK1701109339\r\n"               +
                         "Max-Forwards: 70\r\n"                                                      +
                         "Content-Type: application/sdp\r\n"                                         +
                         "Privacy: none\r\n"                                                         +
                         "P-Preferred-Identity: <sip:hogehoge@example.com>\r\n"                      +
                         "User-Agent: Some User-Agent\r\n"                                           +
                         "Proxy-Authorization: Digest username=\"hogehoge\", realm=\"example.com\"," +
                         " nonce=\"15044921123142536\", uri=\"sip:0312345678@192.168.0.1:5060\","    +
                         " response=\"358a640a266ad4eb3ed82f0746c82dfd\"\r\n"                        +
                         "\r\n"                                                                      +
                         "v=0\r\n" )

    packet := newPacket(forward, garbage)
    sip.ParseUDP(packet)
    assert.Equal(t, 1, store.size(), "There should be one message published.")

    fields:=store.events[0].Fields
    notes:=fields["sip.notes"]
    assert.Contains(t, fmt.Sprintf("%s",notes) ,"Incompleted message","There should be contained." )
}

func TestParseUdp_IncompletePacketInHeader(t *testing.T) {
    store := &eventStore{}
    sip := newSIP(store, testing.Verbose())

    garbage := []byte(  "INVITE sip:0312345678@192.168.0.1:5060 SIP/2.0\r\n"                        +
                         "Via: SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK1701109339\r\n"               +
                         "From: <sip:hogehoge@example.cm>;tag=1451088881\r\n"                        +
                         "To: <sip:0312345678@192.168.0.1>\r\n"                                      +
                         "Call-ID: hogehoge@10.0.0.1\r\n"                                            +
                         "CSeq: 2 INVITE\r\n"                                                        +
                         "Contact: <sip:1833176976@10.0.0.1:5060;transport=udp>\r\n"                 +
                         "Supported: 100rel, timer\r\n"                                              +
                         "Allow: INVITE, ACK, CANCEL, BYE, UPDATE, PRACK\r\n"                        +
                         "Content-Length: 134\r\n"                                                   +
                         "Session-Expires: 180\r\n")

    packet := newPacket(forward, garbage)
    sip.ParseUDP(packet)
    assert.Equal(t, 1, store.size(), "There should be one message published.")

    fields:=store.events[0].Fields
    notes:=fields["sip.notes"]
    assert.Contains(t, fmt.Sprintf("%s",notes) ,"Incompleted message","There should be contained." )
}

func TestParseUdp_compact_form(t *testing.T){
    store := &eventStore{}
    sip := newSIP(store, testing.Verbose())
    garbage := []byte( "INVITE sip:0312345678@192.168.0.1;user=phone SIP/2.0\r\n"   +
                       "Via: SIP/2.0/UDP 10.0.0.3:5060;branch=z9hG4bK81075724\r\n"  +
                       "f: <sip:sipurl@192.168.0.1>;tag=269050131\r\n"              +
                       "t: <sip:0312341234@192.168.0.1;user=phone>\r\n"             +
                       "m: <sip:301234123@10.0.0.1;user=phone>\r\n"                 +
                       "i: hogehoge@192.168.0.1\r\n"                                +
                       "CSeq: 1 INVITE\r\n"                                         +
                       "Max-Forwards: 70\r\n"                                       +
                       "s: Sample Message\r\n"                                      +
                       "e: none\r\n"                                                +
                       "Allow: INVITE, ACK, CANCEL, BYE, UPDATE, PRACK\r\n"         +
                       "k: 100rel,timer\r\n"                                        +
                       "v: SIP/2.0/UDP 10.0.0.2:5060;branch=z9hG4bK81075722\r\n"    +
                       "Session-Expires: 300\r\n"                                   +
                       "Privacy: none\r\n"                                          +
                       "P-Preferred-Identity: <tel:0387654321>\r\n"                 +
                       "c: application/sdp\r\n"                                     +
                       "l: 107\r\n"                                                 +
                       "Via: SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK81075720\r\n"  +
                       "\r\n"                                                       +
                       "v=0\r\n"                                                    +
                       "o=- 0 0 IN IP4 10.0.0.1\r\n"                                +
                       "s=-\r\n"                                                    +
                       "c=IN IP4 10.0.0.1\r\n"                                      +
                       "t=0 0\r\n"                                                  +
                       "m=audio 5012 RTP/AVP 0\r\n"                                 +
                       "a=rtpmap:0 PCMU/8000\r\n")
    packet := newPacket(forward, garbage)
    sip.ParseUDP(packet)
    assert.Equal(t, 1, store.size(), "There should be one message published.")
    if store.size() == 1{
        fields:=store.events[0].Fields
        headers,_:=fields["sip.headers"].(common.MapStr)
        // mandatories
        assert.Equal(t, "INVITE",
                        fields["sip.method"], 
                        "SIP method should be [INVITE].")

                        assert.Equal(t, "sip:0312345678@192.168.0.1;user=phone", 
                        fields["sip.request-uri"], 
                        "Request uri should be [sip:0312345678@192.168.0.1;user=phone].")

        assert.Equal(t, "hogehoge@192.168.0.1" ,
                        fields["sip.call-id"], 
                        "Call-ID should be [hogehoge@192.168.0.1].")

        assert.Equal(t, "<sip:sipurl@192.168.0.1>;tag=269050131",
                        fields["sip.from"], 
                        "From should be [<sip:sipurl@192.168.0.1>;tag=269050131].")

        assert.Equal(t, "<sip:0312341234@192.168.0.1;user=phone>",
                        fields["sip.to"],
                        "To should be [<sip:0312341234@192.168.0.1;user=phone>].")

        assert.Equal(t, "1 INVITE",
                        fields["sip.cseq"],
                        "CSeq should be [1 INVITE].")
        // headers
        assert.Equal(t, "application/sdp",
                        fmt.Sprintf("%s",(headers["content-type"].([]common.NetString))[0]),
                        "Content-type should be [application/sdp].")

        assert.Equal(t, "Sample Message",
                        fmt.Sprintf("%s",(headers["subject"].([]common.NetString))[0]),
                        "Subject should be [Sample Message].")

        assert.Equal(t, "none",
                        fmt.Sprintf("%s",(headers["content-encoding"].([]common.NetString))[0]),
                        "Content-Encoding should be [none].")

        assert.Equal(t, "70",
                        fmt.Sprintf("%s",(headers["max-forwards"].([]common.NetString))[0]),
                        "Max-Forwards should be [70].")

        assert.Contains(t, headers["allow"], common.NetString("INVITE"), "Allow should contain INVITE value.")
        assert.Contains(t, headers["allow"], common.NetString("ACK")   , "Allow should contain ACK value.")
        assert.Contains(t, headers["allow"], common.NetString("CANCEL"), "Allow should contain CANCEL value.")
        assert.Contains(t, headers["allow"], common.NetString("BYE")   , "Allow should contain BYE valu.")
        assert.Contains(t, headers["allow"], common.NetString("UPDATE"), "Allow should contain UPDATE value.")
        assert.Contains(t, headers["allow"], common.NetString("PRACK") , "Allow should contain PRACK value.")

        assert.Contains(t, headers["supported"], common.NetString("100rel"), "Supported should contain 100rel value.")
        assert.Contains(t, headers["supported"], common.NetString("timer") , "Supported should contain timer value.")

        assert.Equal(t, "300",
                        fmt.Sprintf("%s",(headers["session-expires"].([]common.NetString))[0]),
                        "There should be [300].")

        assert.Equal(t, "none",
                        fmt.Sprintf("%s",(headers["privacy"].([]common.NetString))[0]),
                        "There should be [none].")

        assert.Equal(t, "<tel:0387654321>",
                        fmt.Sprintf("%s",(headers["p-preferred-identity"].([]common.NetString))[0]),
                        "There should be [<tel:0387654321>].")

        assert.Equal(t, "107",
                        fmt.Sprintf("%s",(headers["content-length"].([]common.NetString))[0]),
                        "There should be [107].")


        via0:="SIP/2.0/UDP 10.0.0.3:5060;branch=z9hG4bK81075724"
        assert.Equal(t, via0,
                        fmt.Sprintf("%s",(headers["via"].([]common.NetString))[0]),
                        fmt.Sprintf("There should be [%s].",via0))
        via1:="SIP/2.0/UDP 10.0.0.2:5060;branch=z9hG4bK81075722"
        assert.Equal(t, via1,
                        fmt.Sprintf("%s",(headers["via"].([]common.NetString))[1]),
                        fmt.Sprintf("There should be [%s].",via1))
        via2:="SIP/2.0/UDP 10.0.0.1:5060;branch=z9hG4bK81075720"
        assert.Equal(t, via2,
                        fmt.Sprintf("%s",(headers["via"].([]common.NetString))[2]),
                        fmt.Sprintf("There should be [%s].",via2))
    }
}

func TestPaseDetailURI(t *testing.T){
    var uri string
    var user_info string; var host string; var port string; var uri_params []string
    sip:=sipPlugin{}

    uri=`sip:0312341234@10.0.0.1:5060`
    user_info, host, port, uri_params=sip.parseDetailURI(uri)
    assert.Equal(t,"0312341234",   user_info,    "User info should be [0312341234]." )
    assert.Equal(t,"10.0.0.1",     host,         "Host should be [10.0.0.1]." )
    assert.Equal(t,"5060",         port,         "Port should be [5060]." )
    assert.Equal(t,0,              len(uri_params),"Parameter length should be [1]." )

    uri=`sip:0312341234@10.0.0.1:5060;user=phone`
    user_info, host, port, uri_params=sip.parseDetailURI(uri)
    assert.Equal(t,"0312341234",   user_info,    "User info should be [0312341234]." )
    assert.Equal(t,"10.0.0.1",     host,         "Host should be [10.0.0.1]." )
    assert.Equal(t,"5060",         port,         "Port should be [5060]." )
    assert.Equal(t,1,              len(uri_params),"Parameter length should be [1]." )
    assert.Contains(t,uri_params,  "user=phone", "Parameter should have [user=phone]." )

    uri=`tel:+81312341234;user=phone`
    user_info, host, port, uri_params=sip.parseDetailURI(uri)
    assert.Equal(t,"",             user_info,      "User info should be []." )
    assert.Equal(t,"+81312341234", host,           "Host should be [+81312341234]." )
    assert.Equal(t,"",             port,           "Port should be []." )
    assert.Equal(t,1,              len(uri_params),"Parameter length should be [1]." )
    assert.Contains(t,uri_params,  "user=phone", "Parameter should have [user=phone]." )

    uri=`sip:bob:password;npdi=yes;rn=0312341234@10.0.0.1:5060;user=phone;lr;transport=udp;ttl=3;method=INVITE;cpc=test`
    user_info, host, port, uri_params=sip.parseDetailURI(uri)
    assert.Equal(t,"bob:password;npdi=yes;rn=0312341234", user_info, "User info should be [bob:password;npdi=yes;rn=0312341234]." )
    assert.Equal(t,"10.0.0.1",     host,           "Host should be [10.0.0.1]." )
    assert.Equal(t,"5060",         port,           "Port should be [5060]." )
    assert.Equal(t,6,              len(uri_params),"Parameter length should be [1]." )
    assert.Contains(t,uri_params,  "user=phone", "Parameter should have [user=phone]." )
    assert.Contains(t,uri_params,  "lr", "Parameter should have [lr]." )
    assert.Contains(t,uri_params,  "transport=udp", "Parameter should have [transport=udp]." )
    assert.Contains(t,uri_params,  "ttl=3", "Parameter should have [ttl=3]." )
    assert.Contains(t,uri_params,  "method=INVITE", "Parameter should have [method=INVITE]." )
    assert.Contains(t,uri_params,  "cpc=test", "Parameter should have [cpc=test]." )

    uri=`sip:+81333334444;npdi;rn=+81312341234@[2001:ab:fe:3::5]:5060`
    user_info, host, port, uri_params=sip.parseDetailURI(uri)
    assert.Equal(t,"+81333334444;npdi;rn=+81312341234", user_info, "User info should be [+81333334444;npdi;rn=+81312341234]." )
    assert.Equal(t,"[2001:ab:fe:3::5]",     host,           "Host should be [[2001:ab:fe:3::5]]." )
    assert.Equal(t,"5060",         port,           "Port should be [5060]." )
    assert.Equal(t,0,              len(uri_params),"Parameter length should be [1]." )

}

func TestParseDetailNameAddr(t *testing.T) {
    sip := sipPlugin{}
    name_addr:="\"display_name\"<sip:0312341234@10.0.0.1:5060;user=phone>;hogehoge"
    display_name, user_info, host, port, uri_params, params:=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"display_name", display_name, "DisplayName should be [display_name]." )
    assert.Equal(t,"0312341234",   user_info,    "User info should be [0312341234]." )
    assert.Equal(t,"10.0.0.1",     host,         "Host should be [10.0.0.1]." )
    assert.Equal(t,"5060",         port,         "Port should be [5060]." )
    assert.Contains(t,uri_params,  "user=phone", "Parameter should have [user=phone]." )
    assert.Contains(t,params,      "hogehoge",   "Parameter should have [hogehoge]." )

    name_addr="<sip:0312341234@10.0.0.1>"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"",           display_name, "DisplayName should be []." )
    assert.Equal(t,"0312341234", user_info,    "User info should be [0312341234]." )
    assert.Equal(t,"10.0.0.1",   host,         "Host should be [10.0.0.1]." )
    assert.Equal(t,port,         "",           "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,0 , len(params),    "Parameter should not have any value." )

    name_addr="Mr. Watson <sip:0312341234@10.0.0.1>"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"Mr. Watson",           display_name, "DisplayName should be []." )
    assert.Equal(t,"0312341234", user_info,    "User info should be [0312341234]." )
    assert.Equal(t,"10.0.0.1",   host,         "Host should be [10.0.0.1]." )
    assert.Equal(t,port,         "",           "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,0 , len(params),   "Parameter should not have any value." )

    name_addr="\"display_name\"<sip:0312341234@10.0.0.1>"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"display_name", display_name, "DisplayName should be [display_name]." )
    assert.Equal(t,"0312341234",   user_info,    "User info should be [0312341234]." )
    assert.Equal(t,"10.0.0.1",     host,         "Host should be [10.0.0.1]." )
    assert.Equal(t,port,           "",           "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,0 , len(params),   "Parameter should not have any value." )

    name_addr="<sip:whois.this;user=phone>"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"",            display_name, "DisplayName should be []." )
    assert.Equal(t,"",            user_info,    "User info should be []." )
    assert.Equal(t,"whois.this",  host,         "Host should be [whois.this]." )
    assert.Equal(t,"",            port,         "Port should be []." )
    assert.Contains(t,uri_params,"user=phone",  "Parameter should have [user=phone]." )
    assert.Equal(t,0 , len(params),   "Parameter should not have any value." )

    name_addr=" \"0333334444\" <sip:[2001:30:fe::4:123];user=phone >"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"0333334444",          display_name, "DisplayName should be [display_name]." )
    assert.Equal(t,"",                    user_info,    "User info should be []." )
    assert.Equal(t,"[2001:30:fe::4:123]", host,         "Host should be [2001:30:fe::4:123]." )
    assert.Equal(t,"",                    port,         "Port should be [5060]." )
    assert.Contains(t,uri_params,"user=phone",          "Parameter should have [user=phone]." )
    assert.Equal(t, 0 , len(params),   "Parameter should not have any value." )

    name_addr=" \"0333334444\" <sips:user:password@[2001:30:fe::4:123]:5060 ;user=phone>"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"0333334444",          display_name, "DisplayName should be [display_name]." )
    assert.Equal(t,"user:password",       user_info,    "User info should be [user:password]." )
    assert.Equal(t,"[2001:30:fe::4:123]", host,         "Host should be [2001:30:fe::4:123]." )
    assert.Equal(t,"5060",                port,         "Port should be [5060]." )
    assert.Contains(t,uri_params,"user=phone",          "Parameter should have [user=phone]." )
    assert.Equal(t, 0 , len(params),   "Parameter should not have any value." )

    name_addr="\"0312341234\"<tel:+81312341234;user=phone>;tag=1234"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"0312341234",   display_name, "DisplayName should be [0333334444]." )
    assert.Equal(t,"",             user_info,    "User info should be []." )
    assert.Equal(t,"+81312341234", host,         "Host should be [+81312341234]." )
    assert.Equal(t,"",             port,         "Port should be [5060]." )
    assert.Contains(t,uri_params,"user=phone",   "Parameter should have [user=phone]." )
    assert.Contains(t,params,    "tag=1234",     "Parameter should have [user=phone]." )

    name_addr="<tel:+81312341234:5060;user=phone>;tag=1234"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"",             display_name, "DisplayName should be [0333334444]." )
    assert.Equal(t,"",             user_info,    "User info should be []." )
    assert.Equal(t,"+81312341234", host,         "Host should be [+81312341234]." )
    assert.Equal(t,"5060",         port,         "Port should be [5060]." )
    assert.Contains(t,uri_params,"user=phone",   "Parameter should have [user=phone]." )
    assert.Contains(t,params,    "tag=1234",     "Parameter should have [user=phone]." )

    name_addr="<sip:a>;tag=1234"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"",  display_name, "DisplayName should be []." )
    assert.Equal(t,"",  user_info,    "User info should be []." )
    assert.Equal(t,"a", host,         "Host should be [a]." )
    assert.Equal(t,"",  port,         "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Contains(t,params,    "tag=1234",   "Parameter should have [tag=1234]." )

    name_addr="<sip:a>"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"",  display_name, "DisplayName should be []." )
    assert.Equal(t,"",  user_info,    "User info should be []." )
    assert.Equal(t,"a", host,         "Host should be [a]." )
    assert.Equal(t,"",  port,         "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,0 , len(params),   "Parameter should not have any value." )

    name_addr=` "\"Tokyo\" is capital of \"Japan\""  <tel:+81312341234;user=phone >`
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,`\"Tokyo\" is capital of \"Japan\"`,
                                   display_name, `DisplayName should be [\"Tokyo\" is capital of \"Japan\"].` )
    assert.Equal(t,"",             user_info,    "User info should be []." )
    assert.Equal(t,"+81312341234", host,         "Host should be [+81312341234]." )
    assert.Equal(t,"",             port,         "Port should be []." )
    assert.Contains(t,uri_params,  "user=phone",   "Parameter should have [user=phone]." )
    assert.Equal(t,0 , len(params),   "Parameter should not have any value." )

    name_addr=` "<::;@>"  <tel:+81312341234>;`
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,`<::;@>`, display_name, `DisplayName should be [<::;@>].` )
    assert.Equal(t,"",             user_info,    "User info should be []." )
    assert.Equal(t,"+81312341234", host,         "Host should be [+81312341234]." )
    assert.Equal(t,"",             port,         "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Contains(t,params,    "",   "Parameter should have [user=phone]." )


    // malformed case
    name_addr="sip:10.0.0.1"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"", display_name, "DisplayName should be []." )
    assert.Equal(t,"", user_info,    "User info should be []." )
    assert.Equal(t,"", host,         "Host should be []." )
    assert.Equal(t,"", port,         "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,0 , len(params),  "Parameter should not have any value." )

    // malformed case
    name_addr="<10.0.0.1>"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"", display_name, "DisplayName should be []." )
    assert.Equal(t,"", user_info,    "User info should be []." )
    assert.Equal(t,"", host,         "Host should be []." )
    assert.Equal(t,"", port,         "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,0 , len(params),  "Parameter should not have any value." )

    // malformed case
    name_addr="<mail:10.0.0.1>;tag=1234"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"", display_name, "DisplayName should be []." )
    assert.Equal(t,"", user_info,    "User info should be []." )
    assert.Equal(t,"", host,         "Host should be []." )
    assert.Equal(t,"", port,         "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,1 , len(params),  "Parameter should not have any value." )

    // malformed case
    name_addr="<sip:>;tag=1234"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"", display_name, "DisplayName should be []." )
    assert.Equal(t,"", user_info,    "User info should be []." )
    assert.Equal(t,"", host,         "Host should be []." )
    assert.Equal(t,"", port,         "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Contains(t,params,"tag=1234",   "Parameter should have [user=1234]." )

    // malformed case
    name_addr="<sip:>"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"",  display_name, "DisplayName should be []." )
    assert.Equal(t,"",  user_info,    "User info should be []." )
    assert.Equal(t,"", host,         "Host should be [a]." )
    assert.Equal(t,"",  port,         "Port should be []." )
    assert.Equal(t,0 , len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,0 , len(params),   "Parameter should not have any value." )

    // malformed case
    name_addr="\"test\"<sip:>"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"test",display_name,   "DisplayName should be []." )
    assert.Equal(t,"",    user_info,      "User info should be []." )
    assert.Equal(t,"",    host,           "Host should be [a]." )
    assert.Equal(t,"",    port,           "Port should be []." )
    assert.Equal(t,0 ,    len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,0 ,    len(params),    "Parameter should not have any value." )

    // malformed case
    name_addr="\"test\"<>;tag=1234"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"test",  display_name, "DisplayName should be []." )
    assert.Equal(t,"",      user_info,    "User info should be []." )
    assert.Equal(t,"",      host,         "Host should be [a]." )
    assert.Equal(t,"",      port,         "Port should be []." )
    assert.Equal(t,0 ,    len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,1 ,      len(params),   "Parameter should not have any value." )

    // malformed case
    name_addr="<tel:+81312341234:5060>tag=1234"
    display_name, user_info, host, port, uri_params, params=sip.parseDetailNameAddr(name_addr)

    assert.Equal(t,"",             display_name, "DisplayName should be [0333334444]." )
    assert.Equal(t,"",             user_info,    "User info should be []." )
    assert.Equal(t,"+81312341234", host,         "Host should be [+81312341234]." )
    assert.Equal(t,"5060",         port,         "Port should be [5060]." )
    assert.Equal(t,0 ,      len(uri_params),"Parameter should not have any value." )
    assert.Equal(t,0 ,      len(params),   "Parameter should not have any value." )
}

