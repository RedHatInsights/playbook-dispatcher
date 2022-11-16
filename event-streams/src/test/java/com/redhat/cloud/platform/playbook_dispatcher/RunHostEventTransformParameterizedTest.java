package com.redhat.cloud.platform.playbook_dispatcher;

import java.util.Arrays;
import java.util.Collection;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.redhat.cloud.platform.playbook_dispatcher.types.RunHostEvent;

import org.apache.kafka.connect.source.SourceRecord;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.Parameterized;
import org.junit.runners.Parameterized.Parameters;

import static org.junit.Assert.assertEquals;

@RunWith(Parameterized.class)
public class RunHostEventTransformParameterizedTest {

    private RunHostEventTransform<SourceRecord> transform;

    private final SourceRecord record;
    private final String eventType;

    public RunHostEventTransformParameterizedTest(String eventType, SourceRecord record) {
        this.eventType = eventType;
        this.record = record;
    }

    @Parameters(name="{0}")
    public static Collection<Object[]> data() {
        return Arrays.asList(new Object[][] {
            {
                "create",
                Factory.newHostEventCreate()
            },
            {
                "read",
                Factory.newHostEventRead()
            },
            {
                "update",
                Factory.newHostEventUpdate()
            },
            {
                "delete",
                Factory.newHostEventDelete()
            }
        });
    }

    @Before
    public void before() {
        transform = new RunHostEventTransform<>();
        transform.configure(Map.of("topic", "foo.baz", "table", "run_hosts"));
    }

    @After
    public void after() {
        transform.close();
    }

    @Test
    public void testKey() throws Exception {
        final SourceRecord result = transform.apply(this.record);
        assertEquals("7609546c-f965-4c9c-966c-9e15f4ecbc5f", result.key());
    }

    @Test
    public void testHeaders() throws Exception {
        final SourceRecord result = transform.apply(this.record);
        assertEquals(result.headers().lastWithName(RunHostEventTransform.HEADER_EVENT_TYPE).value(), this.eventType);
        assertEquals(result.headers().lastWithName(RunHostEventTransform.HEADER_STATUS).value(), "success");
    }

    @Test
    public void testValue() throws Exception {
        final SourceRecord result = transform.apply(this.record);
        final ObjectMapper mapper = new ObjectMapper();
        final RunHostEvent value = mapper.readValue((String) result.value(), RunHostEvent.class);

        assertEquals(this.eventType, value.getEventType().value());
        assertEquals("7609546c-f965-4c9c-966c-9e15f4ecbc5f", value.getPayload().getId());
        assertEquals("success", value.getPayload().getStatus().toString());
        assertEquals("2021-03-10T08:18:12.370585Z", value.getPayload().getCreatedAt());
        assertEquals("2021-03-10T09:18:12.370585Z", value.getPayload().getUpdatedAt());
    }

    @Test
    public void testTopic() throws Exception {
        final SourceRecord result = transform.apply(this.record);
        assertEquals(result.topic(), "foo.baz");
    }
}
