package vm

import (
	"testing"
)

func TestNamespace(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{`
		module Foo
		  class Bar
		    def bar
		      10
		    end
		  end
		end

		Foo::Bar.new.bar
		`, 10},
		{`
		class Foo
		  class Bar
		    def bar
		      10
		    end
		  end
		end

		Foo::Bar.new.bar
		`, 10},
		{`
		class Foo
		  def bar
		    100
		  end

		  class Bar
		    def bar
		      10
		    end
		  end
		end

		Foo.new.bar + Foo::Bar.new.bar
		`, 110},
		{`
		class Foo
		  def bar
		    100
		  end
		end

		module Baz
		  class Bar
		    def bar
		      Foo.new.bar
		    end
		  end
		end

		Baz::Bar.new.bar
		`, 100},
		{`
		module Baz
		  class Bar
		    class Foo
		      def bar
			100
		      end
		    end
		  end
		end

		Baz::Bar::Foo.new.bar
		`, 100},
		{`
		module Baz
		  class Foo
		    def bar
		      100
		    end
		  end

		  class Bar
		    def bar
		      Foo.new.bar
		    end
		  end
		end

		Baz::Bar.new.bar
		`, 100},
		{`
		module Baz
		  class Bar
		    def bar
		      Foo.new.bar
		    end

		    class Foo
		      def bar
			100
		      end
		    end
		  end
		end

		Baz::Bar.new.bar
		`, 100},
		{`
		module Foo
		  class Bar
		    def bar
		      10
		    end
		  end
		end

		module Baz
		  class Bar < Foo::Bar
		    def foo
		      100
		    end
		  end
		end

		b = Baz::Bar.new
		b.foo + b.bar
		`, 110},
		{`
		module A
		  class B
		    class C
		      class D
		        def e
		          10
		        end
		      end
		    end
		  end
		end

		A::B::C::D.new.e
		`, 10},
		{`
		class Foo
		  def self.bar
		    10
		  end
		end

		Object::Foo.bar
		`, 10},

		{`
		Foo = 10

		Object::Foo
		`, 10},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)

		if isError(evaluated) {
			t.Fatalf("got Error: %s.\n Input %s", evaluated.(*Error).Message, tt.input)
		}

		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestRequireSuccess(t *testing.T) {
	input := `
	require("file")

	File.extname("foo.rb")
	`
	evaluated := testEval(t, input)

	if isError(evaluated) {
		t.Fatalf("got Error: %s", evaluated.(*Error).Message)
	}

	testStringObject(t, evaluated, ".rb")

}

func TestRequireFail(t *testing.T) {
	input := `
	require("bar")
	`
	expected := `Can't require "bar"`

	evaluated := testEval(t, input)

	if !isError(evaluated) {
		t.Fatalf("Should return an error")
	}

	if evaluated.(*Error).Message != expected {
		t.Fatalf("Error message should be '%s'", expected)
	}
}

func TestPrimitiveType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`100.class.name
			`,
			"Integer",
		},
		{
			`Integer.name
			`,
			"Integer",
		},
		{
			`"123".class.name
			`,
			"String",
		},
		{
			`String.name
			`,
			"String",
		},
		{
			`true.class.name
			`,
			"Boolean",
		},
		{
			`Boolean.name
			`,
			"Boolean",
		},
		{
			`
			nil.class.name
			`,
			"Null",
		},
		{
			`
			Integer.name
			`,
			"Integer",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)

		if isError(evaluated) {
			t.Fatalf("got Error: %s", evaluated.(*Error).Message)
		}

		testStringObject(t, evaluated, tt.expected)
	}
}

func TestMethodCallWithoutSelf(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{
			`
			class Foo
			  def set_x(x)
			    @x = x
			  end

			  def foo
			    set_x(10)
			    a = 10
			    @x + a
			  end
			end

			f = Foo.new
			f.foo
			`,
			20,
		},
		{
			`
			class Foo
			  def bar=(x)
			    @bar = x
			  end

			  def bar
			    @bar
			  end
			end

			f = Foo.new
			f.bar = 10
			f.bar
			`,
			10,
		},
		{
			`
			class Foo
			  def set_x(x)
			    @x = x
			  end

			  def foo
			    set_x(10 + 10 * 100)
			    a = 10
			    @x + a
			  end
			end

			f = Foo.new
			f.foo
			`,
			1020,
		},
		{
			`class Foo
				def bar
					10
				end

				def foo
					bar = 100
					10 + bar
				end
			end

			f = Foo.new
			f.foo
			`,
			110,
		},
		{
			`class Foo
				def bar
					10
				end

				def foo
					a = 10
					bar + a
				end
			end

			Foo.new.foo
			`,
			20,
		},
		{
			`class Foo
				def self.bar
					10
				end

				def self.foo
					a = 10
					bar + a
				end
			end

			Foo.foo
			`,
			20,
		},
		{
			`class Foo
				def bar
					100
				end

				def self.bar
					10
				end

				def foo
					a = 10
					bar + a
				end
			end

			Foo.new.foo
			`,
			110,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)

		if isError(evaluated) {
			t.Fatalf("got Error: %s", evaluated.(*Error).Message)
		}

		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestClassMethodEvaluation(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Bar
				def self.foo
					10
				end
			end
			Bar.foo;
			`,
			10,
		},
		{
			`
			class Bar
				def self.foo
					10
				end
			end
			class Foo < Bar; end
			class FooBar < Foo; end
			FooBar.foo
			`,
			10,
		},
		{
			`
			class Foo
				def self.foo
					10
				end
			end

			class Bar < Foo; end
			Bar.foo
			`,
			10,
		},
		{
			`
			class Foo
				def self.foo
					10
				end
			end

			class Bar < Foo
				def self.foo
					100
				end
			end
			Bar.foo
			`,
			100,
		},
		{
			`
			class Bar
				def self.foo
					bar
				end

				def self.bar
					100
				end

				def bar
					1000
				end
			end
			Bar.foo
			`,
			100,
		},
		{
			`
			# Test class method call inside class method.
			class JobPosition
				def initialize(name)
					@name = name
				end

				def self.engineer
					new("Engineer")
				end

				def name
					@name
				end
			end
			job = JobPosition.engineer
			job.name
			`,
			"Engineer",
		},
		{
			`
			class Foo; end
			Foo.new.class.name
			`,
			"Foo",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)

		if isError(evaluated) {
			t.Fatalf("got Error: %s", evaluated.(*Error).Message)
		}

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, expected)
		case string:
			testStringObject(t, evaluated, expected)
		}
	}
}

func TestSelfExpressionEvaluation(t *testing.T) {
	tests := []struct {
		input       string
		expectedObj string
	}{
		{`self`, baseObject},
		{
			`
			class Bar
				def whoami
					self
				end
			end

			Bar.new.whoami;
		`, baseObject},
		{
			`
			class Foo
				Self = self

				def get_self
					Self
				end
			end

			Foo.new.get_self;
			`,
			classObj},
		{
			`
			class Foo
				def class
					Foo
				end
			end

			Foo.new.class
			`,
			classObj},
		{
			`
			class Foo
				def class_name
					self.class.name
				end
			end

			Foo.new.class_name
			`,
			stringObj},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)

		if isError(evaluated) {
			t.Fatalf("got Error: %s", evaluated.(*Error).Message)
		}

		if string(evaluated.objectType()) != tt.expectedObj {
			t.Fatalf("expect self to return %s. got=%s", string(tt.expectedObj), evaluated.objectType())
		}
	}
}

func TestEvalInstanceVariable(t *testing.T) {
	input := `
		class Foo
			def set(x)
				@x = x;
			end

			def get
				@x
			end

			def double_get
				self.get() * 2;
			end
		end

		class Bar
			def set(x)
				@x = x;
			end

			def get
				@x
			end
		end

		f1 = Foo.new
		f1.set(10)

		f2 = Foo.new
		f2.set(20)

		b = Bar.new
		b.set(10)

		f2.double_get() + f1.get() + b.get()
	`

	evaluated := testEval(t, input)

	if isError(evaluated) {
		t.Fatalf("got Error: %s", evaluated.(*Error).Message)
	}

	result, ok := evaluated.(*IntegerObject)

	if !ok {
		t.Errorf("expect result to be an integer. got=%T", evaluated)
	}

	if result.Value != 60 {
		t.Fatalf("expect result to be 60. got=%d", result.Value)
	}
}

func TestEvalInstanceMethodCall(t *testing.T) {
	input := `

		class Bar
			def set(x)
				@x = x
			end
		end

		class Foo < Bar
			def add(x, y)
				x + y
			end
		end

		class FooBar < Foo
			def get
				@x
			end
		end

		fb = FooBar.new
		fb.set(100)
		fb.add(10, fb.get)
	`

	evaluated := testEval(t, input)

	if isError(evaluated) {
		t.Fatalf("got Error: %s", evaluated.(*Error).Message)
	}

	result, ok := evaluated.(*IntegerObject)

	if !ok {
		t.Errorf("expect result to be an integer. got=%T", evaluated)
	}

	if result.Value != 110 {
		t.Errorf("expect result to be 110. got=%d", result.Value)
	}
}

func TestEvalCustomInitializeMethod(t *testing.T) {
	input := `
		class Foo
			def initialize(x, y)
				@x = x
				@y = y
			end

			def bar
				@x + @y
			end
		end

		f = Foo.new(10, 20)
		f.bar
	`

	evaluated := testEval(t, input)

	if isError(evaluated) {
		t.Fatalf("got Error: %s", evaluated.(*Error).Message)
	}

	result, ok := evaluated.(*IntegerObject)

	if !ok {
		t.Errorf("expect result to be an integer. got=%T", evaluated)
	}

	if result.Value != 30 {
		t.Errorf("expect result to be 30. got=%d", result.Value)
	}
}

func TestEvalMethodInheritance(t *testing.T) {
	input := `
		class Foo
			def add(x, y)
				x + y
			end
		end
		Foo.new.add(10, 11)
	`

	evaluated := testEval(t, input)

	if isError(evaluated) {
		t.Fatalf("got Error: %s", evaluated.(*Error).Message)
	}

	result, ok := evaluated.(*IntegerObject)

	if !ok {
		t.Errorf("expect result to be an integer. got=%T", evaluated)
	}

	if result.Value != 21 {
		t.Errorf("expect result to be 21. got=%d", result.Value)
	}
}

func TestEvalClassInheritance(t *testing.T) {
	input := `
		class Bar
		end

		class Foo < Bar
		  def self.add
		    10
		  end
		end

		Foo.superclass.name
	`

	evaluated := testEval(t, input)

	testStringObject(t, evaluated, "Bar")
}

func TestEvalIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			if 10 > 5
				100
			else
				-10
			end
			`,
			100,
		},
		{
			`
			if 5 != 5
				false
			else
				true
			end
			`,
			true,
		},
		{"if true; 10 end", 10},
		{"if false; 10 end", nil},
		{"if 1; 10; end", 10},
		{"if 1 < 2; 10 end", 10},
		{"if 1 > 2; 10 end", nil},
		{"if 1 > 2; 10 else 20 end", 20},
		{"if 1 < 2; 10 else 20 end", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)

		switch tt.expected.(type) {
		case int64:
			testIntegerObject(t, evaluated, tt.expected.(int))
		case bool:
			testBooleanObject(t, evaluated, tt.expected.(bool))
		case nil:
			testNullObject(t, evaluated)
		}

	}
}

func TestEvalPostfix(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1++", 2},
		{"10--", 9},
		{"0--", -1},
		{"-5++", -4},
		{`
		a = 10
		a ++
		`, 11},
		{`
		a = 10
		a --
		`, 9},
		{`
		(1 + 2 * 3)++
		`, 8},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBangPrefixExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!5", false},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestEvalMinusPrefixExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"-5", -5},
		{"-10", -10},
		{"-(-10)", 10},
		{"-(-5)", 5},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestMethodCallWithBlockArgument(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{`
				class Foo
				  def bar
				    yield(1, 3, 5)
				  end
				end

				Foo.new.bar do |first, second, third|
				  first + second * third
				end

				`, 16},
		{`
				class Foo
				  def bar
				    yield
				  end
				end

				Foo.new.bar do
				  3
				end

				`, 3},
		{`
				class Bar
				  def foo
				    yield(10)
				  end
				end

				class Foo
				  def bar
				    yield
				  end
				end

				Bar.new.foo do |num|
				  Foo.new.bar do
				    3 * num
				  end
				end

				`, 30},
		{`
				class Foo
				  def bar
				    0
				  end
				end

				Foo.new.bar do
				  3
				end

				`, 0},
		{`
				class Foo
				  def bar
				    yield
				  end
				end

				i = 10
				Foo.new.bar do
				  i = 3 + i
				end
				i

				`, 13},
		{`
		class Car
		  def initialize
		    yield(self)
		  end

		  def doors=(ds)
		    @doors = ds
		  end

		  def doors
		    @doors
		  end
		end

		car = Car.new do |c|
		  c.doors = 4
		end

		car.doors
				`,
			4},
		{`
		class Foo
		  def bar(x)
		    yield(x)
		  end
		end

		f = Foo.new
		x = 100
		y = 10

		f.bar(10) do |x|
                  y = x + y
		end

		y
		`, 20},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestMethodCallWithNestedBlock(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{`
		class Foo
		  def bar
		    yield
		  end
		end

		a = 100
		i = 10
		b = 1000

		f = Foo.new

		f.bar do
		  i = 3 * a
		  f.bar do
		    i = 3 + i
		  end
		end
		i

		`, 303},
		{`
		class Foo
		  def bar
		    yield
		  end
		end

		i = 10
		a = 100
		b = 1000

		f = Foo.new

		f.bar do
		  a = 20
		  f.bar do
		    b = (3 + i) * a
		  end
		end
		b

		`, 260},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}
