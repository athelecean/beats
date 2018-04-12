### Implementation plan

#### Published for each SIP message(request or response)
- SIP is not a one to one message with request and response. Also order to each message is not determined(a response may be sent after previous response).
- Therefore the SIP response and SIP request is published when packetbeat received the message immidiatory.
- If you need all SIP messages in throughout of SIP dialog, you need to retrieve from Elasticsearch using the SIP Call-ID field etc.

#### Additional timestamp
- Default timestamp field(@timestamp) precision is not sufficient(the sip response is often send immediately when request received eg. 100 Trying).
- Therefore I added the ``sip.unixtimenano``(int64) in order to keep the message order.

#### Request-Line,Status-Line
- In case of SIP request received, stored ``sip.method``(eg.INVITE,BYE,ACK,PRACK) and ``sip.request-uri``.
- In case of SIP response received, stored ``sip.status-code``(eg.200,404) and ``sip.status-phrase``(eg. OK, Ringing)

#### Mandatory headers
- ``sip.from``,``sip.to``,``sip.call-id``,``sip.cseq`` are SIP mandatory headers.

#### SIP Headers
- A SIP header might be exsist multiple lines(eg. Via).
- The description order of the SIP header has a meaning.
- Each SIP header is sotred as dict and thi dict has header values as array.

#### SIP Body
- SIP allowed having mulitple type of body.
- Currently it only supports sdp

#### Raw message
- Recived raw message is stored in ``raw`` field as text value.

#### Sample JSON Output
```json
{
   "_index": "packetbeat-7.0.0-alpha1-2018.01.17",
   "_type": "doc",
   "_id": "14uKBGEBLUdHmvOi5U1L",
   "_score": null,
   "_source": {
     "@timestamp": "2018-01-17T14:34:26.016Z",
     "beat": {
       "name": "Elasticsearch1",
       "hostname": "Elasticsearch1",
       "version": "7.0.0-alpha1"
     },
     "sip.headers": {
       "from": [
         "sipp <sip:sipp@192.168.0.220:5060>;tag=26730SIPpTag003138"
       ],
       "to": [
         "service <sip:service@127.0.0.1:5060>"
       ],
       "cseq": [
         "1 INVITE"
       ],
       "subject": [
         "Performance Test"
       ],
       "contact": [
         "sip:sipp@192.168.0.220:5060"
       ],
       "content-type": [
         "application/sdp"
       ],
       "call-id": [
         "3138-26730@192.168.0.220"
       ],
       "content-length": [
         "137"
       ],
       "via": [
         "SIP/2.0/UDP 192.168.0.220:5060;branch=z9hG4bK-26730-3138-0"
       ],
       "max-forwards": [
         "70"
       ]
     },
     "sip.body": {
       "application/sdp": {
         "o": [
           "user1 53655765 2353687637 IN IP4 192.168.0.220"
         ],
         "s": [
           "-"
         ],
         "c": [
           "IN IP4 192.168.0.220"
         ],
         "t": [
           "0 0"
         ],
         "m": [
           "audio 6000 RTP/AVP 0"
         ],
         "a": [
           "rtpmap:0 PCMU/8000"
         ],
         "v": [
           "0"
         ]
       }
     },
     "sip.request-uri": "sip:service@127.0.0.1:5060",
     "sip.call-id": "3138-26730@192.168.0.220",
     "sip.cseq": "1 INVITE",
     "sip.dst": "127.0.0.1:5060",
     "sip.unixtimenano": 1516199666016756000,
     "type": "sip",
     "sip.method": "INVITE",
     "sip.from": "sipp <sip:sipp@192.168.0.220:5060>;tag=26730SIPpTag003138",
     "sip.to": "service <sip:service@127.0.0.1:5060>",
     "sip.raw": """
INVITE sip:service@127.0.0.1:5060 SIP/2.0
Via: SIP/2.0/UDP 192.168.0.220:5060;branch=z9hG4bK-26730-3138-0
From: sipp <sip:sipp@192.168.0.220:5060>;tag=26730SIPpTag003138
To: service <sip:service@127.0.0.1:5060>
Call-ID: 3138-26730@192.168.0.220
CSeq: 1 INVITE
Contact: sip:sipp@192.168.0.220:5060
Max-Forwards: 70
Subject: Performance Test
Content-Type: application/sdp
Content-Length:   137

v=0
o=user1 53655765 2353687637 IN IP4 192.168.0.220
s=-
c=IN IP4 192.168.0.220
t=0 0
m=audio 6000 RTP/AVP 0
a=rtpmap:0 PCMU/8000

""",
    "sip.src": "192.168.0.220:5060",
    "sip.transport": "udp"
  }
}
```

#### TCP
* ``transport=tcp`` is not supported yet.

#### TODO
* parse detail mode
 - example case from(to)
 - input>> From: "user"<sip:0312341234@bob.com>;tag=zxcvb;otheroption
 - output>
```
{
    "sip.from.raw":"\"user\"<sip:0312341234@bob.com>;tag=zxcvb;otheroption",
    "sip.from.display-name":"user",
    "sip.from.user":"0312341234",
    "sip.from.host":"bob.com",
    "sip.from.param":["tag=zxcvb","otheroption"]
}
```
 - example case cseq
 - input>> CSeq: 1 INVITE 
 - output>
```
{
    "sip.cseq.raw":"1 INVITE",
    "sip.cseq.number":1,
    "sip.cseq.method":"INVITE"
}
```
 - example case request-uri
 - input>> INVITE sip:9012341234;rn=9012340000;npdi=yes@hoge.com:5060;transport=udp;user=phone SIP/2.0
 - output>
```
{
    "sip.request-uri.raw":"sip:9012341234;rn=9012340000;npdi=yes@hoge.com:5060;transport=udp;user=phone"
    "sip.request-uri.user":"9012341234;rn=9012340000;npdi=yes",
    "sip.request-uri.host":"hoge.com",
    "sip.request-uri.port":"5060",
    "sip.request-uri.params":["transport=udp","user=phone"]
}
```
 - example case request-uri(telephone-subscriber)
 - input>> INVITE tel:+819012341234;phone-context=+1234;vnd.company.option=foo SIP/2.0
 - output>
```
{
    "sip.request-uri.raw":"tel:+819012341234;phone-context=+1234;vnd.company.option=foo"
    "sip.request-uri.user":"+819012341234",
    "sip.request-uri.params":["phone-context=+1234","vnd.company.option=foo"]
}
```

* In case of body was encoded, Content-encode
* SIP/TCP
* More body parser.
 - ISUP(SIP-I/SIP-T)
 - multipart/form-data boundary

