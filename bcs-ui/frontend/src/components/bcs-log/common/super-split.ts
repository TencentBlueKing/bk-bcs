const splitString = (str: string, delimiter: string) => {
  const result: string[] = [];

  str.split(delimiter).forEach((str) => {
    result.push(str);
    result.push(delimiter);
  });

  result.pop();

  return result;
};

const splitArray = (ary: string[], delimiter: string) => {
  let result: string[] = [];

  ary.forEach((part) => {
    let subRes: string[] = [];

    part.split(delimiter).forEach((str) => {
      subRes.push(str);
      subRes.push(delimiter);
    });

    subRes.pop();
    subRes = subRes.filter((str) => {
      if (str) {
        return str;
      }
      return undefined;
    });

    result = result.concat(subRes);
  });

  return result;
};

const superSplit = (splittable: string | string[], delimiters: string | string[]): any => {
  if (delimiters.length === 0) {
    return splittable;
  }

  if (typeof splittable === 'string') {
    const delimiter = delimiters[delimiters.length - 1];
    const split = splitString(splittable, delimiter);
    return superSplit(split, delimiters.slice(0, -1));
  }

  if (Array.isArray(splittable)) {
    const delimiter = delimiters[delimiters.length - 1];
    const split = splitArray(splittable, delimiter);
    return superSplit(split, delimiters.slice(0, -1));
  }

  return false;
};

export default superSplit;
