import { Table } from 'bk-magic-vue';

import { TABLE_COLUMNS } from '@/common/constant';


export function setTableColWByMemory(tableInstance: InstanceType<typeof Table>) {
  if (!tableInstance) return;

  const { columns } = tableInstance.store.states;

  const storageStr = localStorage.getItem(TABLE_COLUMNS);
  if (storageStr) {
    let colWidthList = [];
    try {
      colWidthList = JSON.parse(storageStr);
    } catch (error) {
      console.warn('JSON.parse error:', error);
    }

    if (colWidthList.length !== columns.length) {
      updateLocalStorage(columns);
      return;
    }

    colWidthList.forEach((width, index) => {
      columns[index].width = width;
    });

    tableInstance.doLayout();
  } else {
    updateLocalStorage(columns);
  }
}

export function handleHeaderDragend(tableInstance: InstanceType<typeof Table>) {
  if (!tableInstance) return;
  const { columns } = tableInstance.store.states;

  updateLocalStorage(columns);
}

const updateLocalStorage = (columns: any[]) => {
  const columnWidths = columns.map(item => item.realWidth);
  localStorage.setItem(TABLE_COLUMNS, JSON.stringify(columnWidths));
};
