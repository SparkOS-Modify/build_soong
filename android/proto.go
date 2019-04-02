// Copyright 2017 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package android

import (
	"strings"

	"github.com/google/blueprint/proptools"
)

// TODO(ccross): protos are often used to communicate between multiple modules.  If the only
// way to convert a proto to source is to reference it as a source file, and external modules cannot
// reference source files in other modules, then every module that owns a proto file will need to
// export a library for every type of external user (lite vs. full, c vs. c++ vs. java).  It would
// be better to support a proto module type that exported a proto file along with some include dirs,
// and then external modules could depend on the proto module but use their own settings to
// generate the source.

type ProtoFlags struct {
	Flags                 []string
	CanonicalPathFromRoot bool
	Dir                   ModuleGenPath
	SubDir                ModuleGenPath
	OutTypeFlag           string
	OutParams             []string
}

func GetProtoFlags(ctx ModuleContext, p *ProtoProperties) ProtoFlags {
	var protoFlags []string
	if len(p.Proto.Local_include_dirs) > 0 {
		localProtoIncludeDirs := PathsForModuleSrc(ctx, p.Proto.Local_include_dirs)
		protoFlags = append(protoFlags, JoinWithPrefix(localProtoIncludeDirs.Strings(), "-I"))
	}
	if len(p.Proto.Include_dirs) > 0 {
		rootProtoIncludeDirs := PathsForSource(ctx, p.Proto.Include_dirs)
		protoFlags = append(protoFlags, JoinWithPrefix(rootProtoIncludeDirs.Strings(), "-I"))
	}

	return ProtoFlags{
		Flags:                 protoFlags,
		CanonicalPathFromRoot: proptools.BoolDefault(p.Proto.Canonical_path_from_root, true),
		Dir:                   PathForModuleGen(ctx, "proto"),
		SubDir:                PathForModuleGen(ctx, "proto", ctx.ModuleDir()),
	}
}

type ProtoProperties struct {
	Proto struct {
		// Proto generator type.  C++: full or lite.  Java: micro, nano, stream, or lite.
		Type *string `android:"arch_variant"`

		// list of directories that will be added to the protoc include paths.
		Include_dirs []string

		// list of directories relative to the bp file that will
		// be added to the protoc include paths.
		Local_include_dirs []string

		// whether to identify the proto files from the root of the
		// source tree (the original method in Android, useful for
		// android-specific protos), or relative from where they were
		// specified (useful for external/third party protos).
		//
		// This defaults to true today, but is expected to default to
		// false in the future.
		Canonical_path_from_root *bool
	} `android:"arch_variant"`
}

func ProtoRule(ctx ModuleContext, rule *RuleBuilder, protoFile Path, flags ProtoFlags, deps Paths,
	outDir WritablePath, depFile WritablePath, outputs WritablePaths) {

	var protoBase string
	if flags.CanonicalPathFromRoot {
		protoBase = "."
	} else {
		rel := protoFile.Rel()
		protoBase = strings.TrimSuffix(protoFile.String(), rel)
	}

	rule.Command().
		Tool(ctx.Config().HostToolPath(ctx, "aprotoc")).
		FlagWithArg(flags.OutTypeFlag+"=", strings.Join(flags.OutParams, ",")+":"+outDir.String()).
		FlagWithDepFile("--dependency_out=", depFile).
		FlagWithArg("-I ", protoBase).
		Flags(flags.Flags).
		Input(protoFile).
		Implicits(deps).
		ImplicitOutputs(outputs)

	rule.Command().
		Tool(ctx.Config().HostToolPath(ctx, "dep_fixer")).Flag(depFile.String())
}
