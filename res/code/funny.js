console.log('dancing bears!');
function f(config) {
	console.log("funny created");
}

f.prototype.ElementRoot = function() {
	return this.d;
};
f.prototype.Open = function() {
	console.log("application started");
};
system.init.set("example-1.solidcoredata.org/app/ctl/spa/funny", f);
