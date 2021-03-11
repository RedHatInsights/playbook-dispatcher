package com.redhat.cloud.platform.playbook_dispatcher;

import java.util.Arrays;
import java.util.Collection;
import java.util.Collections;
import java.util.Map;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonMappingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.redhat.cloud.platform.playbook_dispatcher.types.RunEvent;

import org.apache.kafka.connect.data.SchemaBuilder;
import org.apache.kafka.connect.data.Struct;
import org.apache.kafka.connect.source.SourceRecord;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.Parameterized;
import org.junit.runners.Parameterized.Parameters;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertTrue;

public class RunEventTransformTest {

    private RunEventTransform<SourceRecord> transform;

    @Before
    public void before() {
        transform = new RunEventTransform<>();
        transform.configure(Map.of("topic", "foo.bar", "table", "runs"));
    }

    @After
    public void after() {
        transform.close();
    }

    @Test
    public void testUnknownTableIsIgnored() throws Exception {
        final Struct key = Factory.getKey();
        final Struct source = new StructBuilder().put("table", "systems").build();
        final Struct value = new StructBuilder()
        .put("after", Factory.getData())
        .put("op", "c")
        .put("source", source)
        .build();

        final SourceRecord record = new SourceRecord(null, null, "public.runs", null, key.schema(), key, value.schema(), value);
        final SourceRecord result = transform.apply(record);
        assertTrue(record == result);
    }

    @Test
    public void testUnknownOperationIsIgnored() throws Exception {
        final Struct key = Factory.getKey();
        final Struct value = new StructBuilder().put("op", "truncate").build();
        final SourceRecord record = new SourceRecord(null, null, null, null, key.schema(), key, value.schema(), value);
        final SourceRecord result = transform.apply(record);
        assertTrue(record == result);
    }

    @Test
    public void testNonStructMessageIgnored() throws Exception {
        final SourceRecord record = new SourceRecord(null, null, null, null, null, null, null, null);
        final SourceRecord result = transform.apply(record);
        assertTrue(record == result);
    }
}
