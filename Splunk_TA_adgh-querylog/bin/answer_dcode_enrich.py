#!/usr/bin/env python3

## BELOW are template for Splunk script

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "lib"))

## Actual code starts here


import json
import base64
from dnslib import DNSRecord, QTYPE, OPCODE, RCODE
from enum import Enum


class ADGHFilteredReason(Enum):
    NotFilteredNotFound = 0
    NotFilteredWhiteList = 1
    NotFilteredError = 2
    FilteredBlackList = 3
    FilteredSafeBrowsing = 4
    FilteredParental = 5
    FilteredInvalid = 6
    FilteredSafeSearch = 7
    FilteredBlockedService = 8
    Rewrite = 9
    RewriteEtcHosts = 10
    RewriteRule = 11


def process_event(querylog):
    querylog_dic = json.loads(querylog)

    # change Elapsed
    querylog_dic["duration"] = int(querylog_dic["Elapsed"] / 100000000)
    del querylog_dic["Elapsed"]

    # may not be split
    # minified_qtypes = ['A', 'AAAA', 'TXT', 'CNAME', 'DNAME', 'SRV', 'NS', 'MX']

    # extract from querylog: "Answer" field
    # decode to message on the wire, instantiate to object
    answer_b64 = querylog_dic["Answer"]
    answer_pkt = base64.b64decode(answer_b64)
    answer_msg = DNSRecord.parse(answer_pkt)

    # extract further fields including:
    #     - [x] | original "Answer" should be deleted after parsed for A,AAAA,TXT,CNAME,DNAME,SRV,NS,MX request
    #     - [x] | others type of request should not
    #     - [x] | create following field: query_type_id, reply_code_id
    #     - [x] | lookup query_type, reply_type
    #     - [x] | find minimum ttl then assign to ttl
    #     - [x] | create transaction_id, transport field

    min_ttl = 4294967295
    tmp_parsed_answer_data = []
    querylog_dic["original_answer"] = str(answer_msg)

    # Transaction and Transport
    querylog_dic["transaction_id"] = answer_msg.header.id
    try:
        querylog_dic["transport"] = querylog_dic["Upstream"].split("://")[0]
        querylog_dic["dest"] = querylog_dic["Upstream"].split("://")[1]
    except KeyError:
        pass
    # source client
    querylog_dic["src_ip"] = querylog_dic["IP"]
    del querylog_dic["IP"]
    # rcode, opcode
    querylog_dic["reply_code_id"] = answer_msg.header.rcode
    querylog_dic["reply_code"] = RCODE.get(answer_msg.header.rcode)
    querylog_dic["query_type_id"] = answer_msg.header.opcode
    querylog_dic["query_type"] = OPCODE.get(answer_msg.header.opcode)
    # answer count
    querylog_dic["answer_count"] = len(answer_msg.rr)
    # query record type
    querylog_dic["record_type"] = QTYPE.get(answer_msg.q.qtype)

    for resp in answer_msg.rr:
        if resp.ttl < min_ttl:
            min_ttl = resp.ttl
        tmp_parsed_answer_data.append(str(resp.rdata))

    del querylog_dic["Answer"]
    querylog_dic["answer"] = " ".join(tmp_parsed_answer_data)

    # extend for reasons of being blocked by adgh
    try:
        querylog_dic["filtered"] = querylog_dic["Result"]["IsFiltered"]
        if querylog_dic["filtered"]:
            querylog_dic["filtered_reason"] = ADGHFilteredReason(querylog_dic["Result"]["Reason"]).name
    except KeyError:
        pass

    del querylog_dic["Result"]
    return querylog_dic


for line in sys.stdin:
    try:
        event = process_event(json.loads(line.strip()))
        print(json.dumps(event))
    except Exception as e:
        pass

