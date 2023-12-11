import { h } from "vue";
import { IReleasedGroup } from "../../../../../../../types/config";
import GROUP_RULE_OPS from '../../../../../../constants/group';

// 已上线分组popover内容
export default (groups: IReleasedGroup[], showTitle: Boolean = true) => {
  return h(
    'div',
    { style: { minWidth: '220px', fontSize: '12px', lineHeight: '16px', color: '#63656e' } },
    [
      showTitle ? h(
        'h3',
        { style: { margin: '0', lineHeight: '20px', color: '#313238' } },
        '已上线分组',
      ) : null,
      h(
        'div',
        {},
        groups.map(group => {
          return h(
            'div',
            {},
            [
              h(
                'div',
                { style: { marginTop: '12px' } },
                group.name
              ),
              h(
                'div',
                { style: { display: 'inline-block', marginTop: '4px', padding: '8px', background: '#f5f7fa' } },
                group.new_selector.labels_and.map((rule, index) => {
                  let opName = '';
                  const op = GROUP_RULE_OPS.find(item => item.id === rule.op);
                  if (op) {
                    opName = op.name;
                  }
                  const valueText = ['in', 'nin'].includes(rule.op) ? `(${(rule.value as string[]).join(', ')})` : rule.value;
                  return h(
                    'span',
                    {
                      rule,
                      class: 'rule-item'
                    },
                    `${index > 0 ? ' & ' : ''}${rule.key}${opName}${valueText}`
                  );
                })
              )
            ]
          );
        })
      )
    ]
  )
};
