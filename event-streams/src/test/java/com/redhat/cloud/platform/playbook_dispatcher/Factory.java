package com.redhat.cloud.platform.playbook_dispatcher;

import com.redhat.cloud.platform.playbook_dispatcher.types.RunEvent.EventType;

import org.apache.kafka.connect.data.Struct;
import org.apache.kafka.connect.source.SourceRecord;

class Factory {

    public static SourceRecord newRecord(Struct key, Struct value) {
        return new SourceRecord(null, null, null, null, null, key, null, value);
    }


    public static Struct getData(String status) {
        return new StructBuilder()
        .put("id", "b5c80cd3-8849-46a2-97e2-368cf62a1cda")
        .put("account", "0000001")
        .put("org_id", "0000001-test")
        .put("recipient", "dd018b96-da04-4651-84d1-187fa5c23f6c")
        .put("correlation_id", "97b04495-68f0-4a41-93b9-d239c0a59b4f")
        .put("url", "http://example.com")
        .put("labels", "{\"foo\": \"bar\"}")
        .put("playbook_name", "test playbook")
        .put("playbook_run_url", "http://example.com")
        .put("sat_id", "16372e6f-1c18-4cdb-b780-50ab4b88e74b")
        .put("sat_org_id", "6826")
        .put("status", status)
        .put("events", "[]")
        .put("created_at", "2021-03-10T08:18:12.370585Z")
        .put("updated_at", "2021-03-10T09:18:12.370585Z")
        .put("timeout", 3600)
        .put("service", "test")
        .build();
    }

    public static Struct getHostData(String status) {
        return new StructBuilder()
        .put("id", "7609546c-f965-4c9c-966c-9e15f4ecbc5f")
        .put("run_id", "f0705502-6049-461f-99f9-0e18846d8222")
        .put("inventory_id", "48e144d2-50c6-4886-8044-7e0791603d97")
        .put("host", "48e144d2-50c6-4886-8044-7e0791603d97")
        .put("status", status)
        .put("log", "")
        .put("events", "[]")
        .put("created_at", "2021-03-10T08:18:12.370585Z")
        .put("updated_at", "2021-03-10T09:18:12.370585Z")
        .put("sat_sequence", 10)
        .build();
    }

    public static Struct getData() {
        return getData("success");
    }

    public static Struct getHostData() {
        return getHostData("success");
    }

    public static Struct getSource() {
        return new StructBuilder()
        .put("table", "runs")
        .build();
    }

    public static Struct getHostSource() {
        return new StructBuilder()
        .put("table", "run_hosts")
        .build();
    }

    public static Struct getKey() {
        return new StructBuilder()
        .put("id", "b5c80cd3-8849-46a2-97e2-368cf62a1cda")
        .build();
    }

    public static Struct getHostKey() {
        return new StructBuilder()
        .put("id", "7609546c-f965-4c9c-966c-9e15f4ecbc5f")
        .build();
    }

    public static SourceRecord newEventCreate() {
        final Struct key = getKey();
        final Struct value = new StructBuilder()
        .put("after", getData())
        .put("op", "c")
        .put("source", getSource())
        .build();

        return new SourceRecord(null, null, "public.runs", null, key.schema(), key, value.schema(), value);
    }

    public static SourceRecord newHostEventCreate() {
        final Struct key = getHostKey();
        final Struct value = new StructBuilder()
        .put("after", getHostData())
        .put("op", "c")
        .put("source", getHostSource())
        .build();

        return new SourceRecord(null, null, "public.run_hosts", null, key.schema(), key, value.schema(), value);
    }

    public static SourceRecord newEventRead() {
        final Struct key = getKey();
        final Struct value = new StructBuilder()
        .put("after", getData())
        .put("op", "r")
        .put("source", getSource())
        .build();

        return new SourceRecord(null, null, "public.runs", null, key.schema(), key, value.schema(), value);
    }

    public static SourceRecord newHostEventRead() {
        final Struct key = getHostKey();
        final Struct value = new StructBuilder()
        .put("after", getHostData())
        .put("op", "r")
        .put("source", getHostSource())
        .build();

        return new SourceRecord(null, null, "public.run_hosts", null, key.schema(), key, value.schema(), value);
    }

    public static SourceRecord newEventUpdate() {
        final Struct key = getKey();
        final Struct value = new StructBuilder()
        .put("before", getData("running"))
        .put("after", getData())
        .put("op", "u")
        .put("source", getSource())
        .build();

        return new SourceRecord(null, null, "public.runs", null, key.schema(), key, value.schema(), value);
    }

    public static SourceRecord newHostEventUpdate() {
        final Struct key = getHostKey();
        final Struct value = new StructBuilder()
        .put("before", getHostData("running"))
        .put("after", getHostData())
        .put("op", "u")
        .put("source", getHostSource())
        .build();

        return new SourceRecord(null, null, "public.run_hosts", null, key.schema(), key, value.schema(), value);
    }

    public static SourceRecord newEventDelete() {
        final Struct key = getKey();
        final Struct value = new StructBuilder()
        .put("before", getData())
        .put("op", "d")
        .put("source", getSource())
        .build();

        return new SourceRecord(null, null, "public.runs", null, key.schema(), key, value.schema(), value);
    }

    public static SourceRecord newHostEventDelete() {
        final Struct key = getHostKey();
        final Struct value = new StructBuilder()
        .put("before", getHostData())
        .put("op", "d")
        .put("source", getHostSource())
        .build();

        return new SourceRecord(null, null, "public.run_hosts", null, key.schema(), key, value.schema(), value);
    }
}
