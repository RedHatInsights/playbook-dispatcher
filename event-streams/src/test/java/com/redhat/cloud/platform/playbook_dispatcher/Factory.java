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
        .put("recipient", "dd018b96-da04-4651-84d1-187fa5c23f6c")
        .put("correlation_id", "97b04495-68f0-4a41-93b9-d239c0a59b4f")
        .put("url", "http://example.com")
        .put("labels", "{\"foo\": \"bar\"}")
        .put("name", "test playbook")
        .put("web_console_url", "http://example.com")
        .put("recipient_config", "{\"sat_id\": \"16372e6f-1c18-4cdb-b780-50ab4b88e74b\", \"sat_org_id\": \"6826\"}")
        .put("status", status)
        .put("events", "[]")
        .put("created_at", "2021-03-10T08:18:12.370585Z")
        .put("updated_at", "2021-03-10T09:18:12.370585Z")
        .put("timeout", 3600)
        .put("service", "test")
        .build();
    }

    public static Struct getData() {
        return getData("success");
    }

    public static Struct getSource() {
        return new StructBuilder()
        .put("table", "runs")
        .build();
    }

    public static Struct getKey() {
        return new StructBuilder()
        .put("id", "b5c80cd3-8849-46a2-97e2-368cf62a1cda")
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

    public static SourceRecord newEventRead() {
        final Struct key = getKey();
        final Struct value = new StructBuilder()
        .put("after", getData())
        .put("op", "r")
        .put("source", getSource())
        .build();

        return new SourceRecord(null, null, "public.runs", null, key.schema(), key, value.schema(), value);
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

    public static SourceRecord newEventDelete() {
        final Struct key = getKey();
        final Struct value = new StructBuilder()
        .put("before", getData())
        .put("op", "d")
        .put("source", getSource())
        .build();

        return new SourceRecord(null, null, "public.runs", null, key.schema(), key, value.schema(), value);
    }
}
