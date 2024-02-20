
-- Create table for event data report
create TABLE IF NOT EXISTS public.event_data_report (
    id SERIAL,
    event_id bigint,
    event_name VARCHAR(70),
    case_type_id VARCHAR(255),
    reference VARCHAR(255),
    field_name VARCHAR(255),
    change_type VARCHAR(255),
    old_record text,
    new_record text,
    previous_event_created_date timestamp,
    event_created_date timestamp,
    analyze_result_detail text,
    potential_risk BOOLEAN NOT NULL DEFAULT FALSE
);
alter sequence public.event_data_report_id_seq OWNED BY public.event_data_report.id CACHE 50;
