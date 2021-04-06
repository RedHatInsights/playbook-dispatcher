package com.redhat.cloud.platform.playbook_dispatcher;

import java.lang.invoke.MethodHandles;
import java.net.URI;
import java.util.Map;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.redhat.cloud.platform.playbook_dispatcher.types.RunEvent;
import com.redhat.cloud.platform.playbook_dispatcher.types.Payload.Status;
import com.redhat.cloud.platform.playbook_dispatcher.types.RunEvent.EventType;
import com.redhat.cloud.platform.playbook_dispatcher.types.Labels;
import com.redhat.cloud.platform.playbook_dispatcher.types.Payload;

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
 * Transforms debezium events to match schema/run.event.yaml
 */
public class RunEventTransform<T extends ConnectRecord<T>> implements Transformation<T> {

    static final String HEADER_EVENT_TYPE = "event_type";
    static final String HEADER_SERVICE = "service";
    static final String HEADER_STATUS = "status";
    static final String HEADER_ACCOUNT = "account";

    private static final String CONFIG_TOPIC = "topic";
    private static final String CONFIG_TABLE = "table";

    private static final String HEARTBEAT_TOPIC_PREFIX = "__debezium-heartbeat-pd";

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
        LOG.info("RunEventTransform configuration: {}", cfg);

        final AbstractConfig config = new SimpleConfig(config(), cfg);
        this.topic = config.getString(CONFIG_TOPIC);
        this.table = config.getString(CONFIG_TABLE);

        LOG.info("RunEventTransform ready");
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

        final Struct value = Utils.<Struct>cast(record.value());

        if (opToEventType(value.getString("op")) == null) {
            return record;
        }

        if (!this.table.equals(value.getStruct("source").getString("table"))) {
            return record;
        }

        final String newKey = this.transformKey(Utils.<Struct>cast(record.key()));
        final RunEvent event = this.transformValue(value);
        final Headers headers = this.createHeaders(event);

        try {
            final String marshalledValue = objectMapper.writeValueAsString(event);
            LOG.info("processed message; key: {}, event_type: {}, service: {}", newKey, event.getEventType(), event.getPayload().getService());
            return record.newRecord(this.topic, record.kafkaPartition(), null, newKey, null, marshalledValue, record.timestamp(), headers);
        } catch (JsonProcessingException e) {
            LOG.error("Error marshalling JSON", e);
            throw new ConnectException("Error marshalling JSON", e);
        }
    }

    private Headers createHeaders(RunEvent event) {
        return new ConnectHeaders()
            .addString(HEADER_EVENT_TYPE, event.getEventType().value())
            .addString(HEADER_SERVICE, event.getPayload().getService())
            .addString(HEADER_STATUS, event.getPayload().getStatus().value())
            .addString(HEADER_ACCOUNT, event.getPayload().getAccount());
    }

    private String transformKey(Struct key) {
        return key.getString("id");
    }

    private RunEvent transformValue(Struct value) {
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

        final Payload payload = this.buildRunPayload(data);
        return newRunEvent(payload, eventType);
    }

    Payload buildRunPayload(Struct input) {
        final Payload payload = new Payload();
        payload.setId(input.getString("id"));
        payload.setAccount(input.getString("account"));
        payload.setRecipient(input.getString("recipient"));
        payload.setCorrelationId(input.getString("correlation_id"));
        payload.setService(input.getString("service"));
        payload.setUrl(URI.create(input.getString("url")));
        payload.setStatus(Status.fromValue(input.getString("status")));
        payload.setTimeout(input.getInt32("timeout"));
        payload.setCreatedAt(input.getString("created_at"));
        payload.setUpdatedAt(input.getString("updated_at"));

        Labels labels;
        try {
            labels = this.objectMapper.readValue(input.getString("labels"), Labels.class);
        } catch (JsonProcessingException e) {
            LOG.warn("Ignoring message labels due to parsing error, id={}, labels={}", input.getString("id"), input.getString("labels"));
            labels = new Labels();
        }

        payload.setLabels(labels);

        return payload;
    }

    @Override
    public void close() {
    }

    static RunEvent newRunEvent(Payload payload, EventType eventType) {
        final RunEvent event = new RunEvent();
        event.setEventType(eventType);
        event.setPayload(payload);
        return event;
    }

    static EventType opToEventType(String operation) {
        switch (operation) {
            case "c":
                return RunEvent.EventType.CREATE;
            case "r":
                return RunEvent.EventType.READ;
            case "u":
                return RunEvent.EventType.UPDATE;
            case "d":
                return RunEvent.EventType.DELETE;
            default:
                return null;
        }
    }

}
