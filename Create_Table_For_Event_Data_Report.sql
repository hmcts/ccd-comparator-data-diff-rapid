
-- Create table for event data report
create TABLE IF NOT EXISTS public.event_data_report (
    event_id bigint,
    reference VARCHAR(255),
    event_name VARCHAR(70),
    previous_event_name VARCHAR(70),
    field_name VARCHAR(255),
    old_record text,
    new_record text,
    array_change_record text,
    analyze_result_detail text,
    event_delta bigint,
    event_created_date timestamp,
    previous_event_created_date timestamp,
    previous_event_user_id varchar(64),
    previous_event_id bigint,
    event_user_id varchar(64),
    change_type VARCHAR(255),
    rule_matched BOOLEAN NOT NULL DEFAULT FALSE,
    case_type_id VARCHAR(255),
    id SERIAL
);
alter sequence public.event_data_report_id_seq OWNED BY public.event_data_report.id CACHE 50;
