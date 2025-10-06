const a = parseInt(process.env.A || '0');
const b = parseInt(process.env.B || '0');
const sum = a + b;

console.log(`${a} + ${b} = ${sum}`);
console.log(JSON.stringify({ a, b, sum }));
