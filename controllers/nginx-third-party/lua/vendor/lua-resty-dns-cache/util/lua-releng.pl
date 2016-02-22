#!/usr/bin/env perl

use strict;
use warnings;

sub file_contains ($$);

my $version;
for my $file (map glob, qw{ *.lua lib/*.lua lib/*/*.lua lib/*/*/*.lua lib/*/*/*/*.lua lib/*/*/*/*/*.lua }) {


    print "Checking use of Lua global variables in file $file ...\n";
    system("luac -p -l $file | grep ETGLOBAL | grep -vE 'require|type|tostring|error|ngx|ndk|jit|setmetatable|getmetatable|string|table|io|os|print|tonumber|math|pcall|xpcall|unpack|pairs|ipairs|assert|module|package|coroutine|[gs]etfenv|next|rawget|rawset|rawlen'");
    file_contains($file, "attempt to write to undeclared variable");
    #system("grep -H -n -E --color '.{81}' $file");
}

sub file_contains ($$) {
    my ($file, $regex) = @_;
    open my $in, $file
        or die "Cannot open $file fo reading: $!\n";
    my $content = do { local $/; <$in> };
    close $in;
    #print "$content";
    return scalar ($content =~ /$regex/);
}

if (-d 't') {
    for my $file (map glob, qw{ t/*.t t/*/*.t t/*/*/*.t }) {
        system(qq{grep -H -n --color -E '\\--- ?(ONLY|LAST)' $file});
    }
}
