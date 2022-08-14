import fs from "node:fs";
import os from 'node:os';
import url from 'node:url';
import chalk from 'chalk';

const __dirname = url.fileURLToPath(new URL('.', import.meta.url));

const art = fs.readFileSync(__dirname + "/ascii.art").toString();

console.error("\n");
console.error(chalk.blue.bold(art));
console.error("\n");
console.error(chalk.blueBright.bold("Rapidly iterate with NodeJS and Containers."));
console.error("\n");
console.log(chalk.greenBright.bold(`platform: ${os.platform}/${os.arch}/${os.release()}`));
console.log(chalk.greenBright.bold(`cpus: ${os.cpus().length}`));
console.log(chalk.greenBright.bold(`endianness: ${os.endianness()}`));
console.log(chalk.greenBright.bold(`free memory: ${(os.freemem() / 1024 / 1024 / 1024).toFixed(2)}GB`));
console.log(chalk.greenBright.bold(`priority: ${os.getPriority()}`));
console.log(chalk.greenBright.bold(`uptime: ${hhmmss(os.uptime())}`));



function hhmmss(seconds) {
    const d = new Date();
    d.setSeconds(seconds);
    return d.toISOString().substr(11, 8);
}