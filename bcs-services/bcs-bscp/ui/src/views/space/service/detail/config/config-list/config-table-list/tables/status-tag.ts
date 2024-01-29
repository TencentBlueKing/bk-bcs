import { h } from 'vue';
import { CONFIG_STATUS_MAP } from '../../../../../../../../constants/config';

const StatusTag = (props: { status: string }) => {
  if (!props.status || props.status === 'UNCHANGE') {
    return '--';
  }
  const tag = CONFIG_STATUS_MAP[props.status as keyof typeof CONFIG_STATUS_MAP];
  return h(
    'span',
    {
      class: ['status-tag', props.status.toLocaleLowerCase()],
      style: {
        color: tag.color,
        backgroundColor: tag.bgColor,
        padding: '4px 10px',
        borderRadius: '2px',
        fontSize: '12px',
      },
    },
    tag.text,
  );
};

export default StatusTag;
