import { WebTracerProvider } from "@opentelemetry/sdk-trace-web";
import { BatchSpanProcessor } from "@opentelemetry/sdk-trace-base";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";

import { DocumentLoadInstrumentation } from "@opentelemetry/instrumentation-document-load";
import { FetchInstrumentation } from "@opentelemetry/instrumentation-fetch";
import { UserInteractionInstrumentation } from "@opentelemetry/instrumentation-user-interaction";

const exporter = new OTLPTraceExporter({
  url: "http://localhost:4318/v1/traces",
  headers: {},
});

const provider = new WebTracerProvider();
provider.addSpanProcessor(new BatchSpanProcessor(exporter));
provider.register();

new DocumentLoadInstrumentation();
new FetchInstrumentation();
new UserInteractionInstrumentation();
