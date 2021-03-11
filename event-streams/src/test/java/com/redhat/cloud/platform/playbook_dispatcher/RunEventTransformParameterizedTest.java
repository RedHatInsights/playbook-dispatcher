package com.redhat.cloud.platform.playbook_dispatcher;

import java.util.Arrays;
import java.util.Collection;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.redhat.cloud.platform.playbook_dispatcher.types.RunEvent;

import org.apache.kafka.connect.source.SourceRecord;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.Parameterized;
import org.junit.runners.Parameterized.Parameters;

import static org.junit.Assert.assertEquals;

@RunWith(Parameterized.class)
public class RunEventTransformParameterizedTest {

    private RunEventTransform<SourceRecord> transform;

    private final SourceRecord record;
    private final String eventType;

    public RunEventTransformParameterizedTest(String eventType, SourceRecord record) {
        this.eventType = eventType;
        this.record = record;
    }

    @Parameters(name="{0}")
    public static Collection<Object[]> data() {
        return Arrays.asList(new Object[][] {
            {
                "create",
                Factory.newEventCreate()
            },
            {
                "read",
                Factory.newEventRead()
            },
            {
                "update",
                Factory.newEventUpdate()
            },
            {
                "delete",
                Factory.newEventDelete()
            }
        });
    }

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
    public void testKey() throws Exception {
        final SourceRecord result = transform.apply(this.record);
        assertEquals("b5c80cd3-8849-46a2-97e2-368cf62a1cda", result.key());
    }

    @Test
    public void testHeaders() throws Exception {
        final SourceRecord result = transform.apply(this.record);
        assertEquals(result.headers().lastWithName(RunEventTransform.HEADER_EVENT_TYPE).value(), this.eventType);
        assertEquals(result.headers().lastWithName(RunEventTransform.HEADER_ACCOUNT).value(), "0000001");
        assertEquals(result.headers().lastWithName(RunEventTransform.HEADER_SERVICE).value(), "test");
        assertEquals(result.headers().lastWithName(RunEventTransform.HEADER_STATUS).value(), "success");
    }

    @Test
    public void testValue() throws Exception {
        final SourceRecord result = transform.apply(this.record);
        final ObjectMapper mapper = new ObjectMapper();
        final RunEvent value = mapper.readValue((String) result.value(), RunEvent.class);

        assertEquals(this.eventType, value.getEventType().value());
        assertEquals("b5c80cd3-8849-46a2-97e2-368cf62a1cda", value.getPayload().getId());
        assertEquals("0000001", value.getPayload().getAccount());
        assertEquals("dd018b96-da04-4651-84d1-187fa5c23f6c", value.getPayload().getRecipient());
        assertEquals("97b04495-68f0-4a41-93b9-d239c0a59b4f", value.getPayload().getCorrelationId());
        assertEquals("test", value.getPayload().getService());
        assertEquals("http://example.com", value.getPayload().getUrl().toString());
        assertEquals("success", value.getPayload().getStatus().toString());
        assertEquals((Object) 3600, value.getPayload().getTimeout());
        assertEquals("2021-03-10T08:18:12.370585Z", value.getPayload().getCreatedAt());
        assertEquals("2021-03-10T09:18:12.370585Z", value.getPayload().getUpdatedAt());
        assertEquals("bar", value.getPayload().getLabels().getAdditionalProperties().get("foo"));
    }

    @Test
    public void testTopic() throws Exception {
        final SourceRecord result = transform.apply(this.record);
        assertEquals(result.topic(), "foo.bar");
    }
}
