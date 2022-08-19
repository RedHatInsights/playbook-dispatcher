package com.redhat.cloud.platform.playbook_dispatcher;

import java.lang.invoke.MethodHandles;
import java.util.Map;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.redhat.cloud.platform.playbook_dispatcher.types.RunHostEvent;
import com.redhat.cloud.platform.playbook_dispatcher.types.HostPayload.Status;
import com.redhat.cloud.platform.playbook_dispatcher.types.RunHostEvent.EventType;
import com.redhat.cloud.platform.playbook_dispatcher.types.HostPayload;

import org.apache.kafka.common.config.AbstractConfig;
import org.apache.kafka.common.config.ConfigDef;
import org.apache.kafka.connect.connector.ConnectRecord;
import org.apache.kafka.connect.data.Struct;
import org.apache.kafka.connect.errors.ConnectException;
import org.apache.kafka.connect.header.ConnectHeaders;
import org.apache.kafka.connect.header.Headers;
import org.apache.kafka.connect.transforms.Transformation;
import org.apache.kafka.connect.transforms.util.SimpleConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Transforms debezium events to match schema/run.host.event.yaml
 */
public class RunHostEventTransform<T extends ConnectRecord<T>> implements Transformation<T> {

    static final String HEADER_EVENT_TYPE = "event_type";
    static final String HEADER_STATUS = "status";

    private static final String CONFIG_TOPIC = "topic";
    private static final String CONFIG_TABLE = "table";

    private static final String HEARTBEAT_TOPIC_PREFIX = "__debezium-heartbeat-pd";
    private static final String HEARTBEAT_ID = "98875b33-b37e-4c35-be8b-d74f321bac28";

    private static final Logger LOG = LoggerFactory.getLogger(MethodHandles.lookup().lookupClass());

    private final ObjectMapper objectMapper = new ObjectMapper();

    private String topic;
    private String table;

    @Override
    public ConfigDef config() {
        return new ConfigDef()
            .define(CONFIG_TOPIC, ConfigDef.Type.STRING, ConfigDef.Importance.HIGH, "Name of the topic to write transformed messages to")
            .define(CONFIG_TABLE, ConfigDef.Type.STRING, ConfigDef.Importance.HIGH, "Name of the table to transform");
    }

    @Override
    public void configure(Map<String, ?> cfg) {
        LOG.info("RunHostEventTransform configuration: {}", cfg);

        final AbstractConfig config = new SimpleConfig(config(), cfg);
        this.topic = config.getString(CONFIG_TOPIC);
        this.table = config.getString(CONFIG_TABLE);

        LOG.info("RunHostEventTransform ready");
    }

    @Override
    public T apply(T record) {
        if (record.topic() != null && record.topic().startsWith(HEARTBEAT_TOPIC_PREFIX)) {
            LOG.info("Received heartbeat");
            return null;
        }

        if (!(record.key() instanceof Struct) || !(record.value() instanceof Struct) || record.valueSchema().field("op") == null) {
            return record;
        }

        final Struct key = Utils.<Struct>cast(record.key());
        if (HEARTBEAT_ID.equals(key.get("id"))) {
            LOG.info("Received heartbeat");
            return null;
        }


        final Struct value = Utils.<Struct>cast(record.value());

        if (opToEventType(value.getString("op")) == null) {
            return record;
        }

        if (!this.table.equals(value.getStruct("source").getString("table"))) {
            return record;
        }

        final String newKey = this.transformKey(key);
        final RunHostEvent event = this.transformValue(value);
        final Headers headers = this.createHeaders(event);

        try {
            final String marshalledValue = objectMapper.writeValueAsString(event);
            LOG.info("processed message; key: {}, event_type: {}", newKey, event.getEventType());
            return record.newRecord(this.topic, record.kafkaPartition(), null, newKey, null, marshalledValue, record.timestamp(), headers);
        } catch (JsonProcessingException e) {
            LOG.error("Error marshalling JSON", e);
            throw new ConnectException("Error marshalling JSON", e);
        }
    }

    private Headers createHeaders(RunHostEvent event) {
        return new ConnectHeaders()
            .addString(HEADER_EVENT_TYPE, event.getEventType().value())
            .addString(HEADER_STATUS, event.getHostPayload().getStatus().value());
    }

    private String transformKey(Struct key) {
        return key.getString("id");
    }

    private RunHostEvent transformValue(Struct value) {
        final String operation = value.getString("op");
        final EventType eventType = opToEventType(operation);

        Struct data;

        if (eventType == EventType.CREATE || eventType == EventType.READ || eventType == EventType.UPDATE) {
            data = value.getStruct("after");
        } else if (eventType == EventType.DELETE) {
            data = value.getStruct("before");
        } else {
            LOG.info("ignoring op: {}", operation);
            return null;
        }

        final HostPayload payload = this.buildRunHostPayload(data);
        return newRunHostEvent(payload, eventType);
    }

    HostPayload buildRunHostPayload(Struct input) {
        final HostPayload payload = new HostPayload();
        payload.setId(input.getString("id"));
        payload.setRunId(input.getString("run_id"));
        payload.setInventoryId(input.getString("inventory_id"));
        payload.setHost(input.getString("host"));
        payload.setLog(input.getString("log"));
        payload.setStatus(Status.fromValue(input.getString("status")));
        payload.setCreatedAt(input.getString("created_at"));
        payload.setUpdatedAt(input.getString("updated_at"));

        if (input.get("sat_sequence") != null) {
            payload.setSatSequence(input.getInt32("sat_sequence"));
        }

        return payload;
    }

    @Override
    public void close() {
    }

    static RunHostEvent newRunHostEvent(HostPayload payload, EventType eventType) {
        final RunHostEvent event = new RunHostEvent();
        event.setEventType(eventType);
        event.setHostPayload(payload);
        return event;
    }

    static EventType opToEventType(String operation) {
        switch (operation) {
            case "c":
                return RunHostEvent.EventType.CREATE;
            case "r":
                return RunHostEvent.EventType.READ;
            case "u":
                return RunHostEvent.EventType.UPDATE;
            case "d":
                return RunHostEvent.EventType.DELETE;
            default:
                return null;
        }
    }

}
