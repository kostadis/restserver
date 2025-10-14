import unittest
import os
from string_replacer import replace_strings, main
import sys
from io import StringIO

class TestStringReplacer(unittest.TestCase):
    def setUp(self):
        self.test_file = "test_file.txt"
        self.output_file = "output_file.txt"
        self.config_file = "config.txt"
        with open(self.test_file, "w") as f:
            f.write("hello world, this is a test. hello again.")
        with open(self.config_file, "w") as f:
            f.write("goodbye:hello,test\n")
            f.write("universe:world\n")

    def tearDown(self):
        if os.path.exists(self.test_file):
            os.remove(self.test_file)
        if os.path.exists(self.output_file):
            os.remove(self.output_file)
        if os.path.exists(self.config_file):
            os.remove(self.config_file)

    def test_replace_strings_inplace(self):
        replace_strings(self.test_file, ["goodbye:hello,test"])
        with open(self.test_file, "r") as f:
            content = f.read()
        self.assertEqual(content, "goodbye world, this is a goodbye. goodbye again.")

    def test_multiple_substitutions_inplace(self):
        with open(self.test_file, "w") as f:
            f.write("apple banana orange apple")
        replace_strings(self.test_file, ["fruit:apple", "yellow:banana"])
        with open(self.test_file, "r") as f:
            content = f.read()
        self.assertEqual(content, "fruit yellow orange fruit")

    def test_replace_strings_to_output_file(self):
        original_content = "apple banana orange apple"
        with open(self.test_file, "w") as f:
            f.write(original_content)
        replace_strings(self.test_file, ["fruit:apple"], self.output_file)
        with open(self.output_file, "r") as f:
            output_content = f.read()
        self.assertEqual(output_content, "fruit banana orange fruit")
        with open(self.test_file, "r") as f:
            input_content = f.read()
        self.assertEqual(input_content, original_content)

    def test_main_with_config_file(self):
        old_argv = sys.argv
        sys.argv = ["string_replacer.py", "-f", self.test_file, "-c", self.config_file]

        # Capture stdout to prevent it from printing to the console during tests
        captured_output = StringIO()
        sys.stdout = captured_output

        main()

        sys.stdout = sys.__stdout__  # Restore stdout
        sys.argv = old_argv

        with open(self.test_file, "r") as f:
            content = f.read()

        self.assertEqual(content, "goodbye universe, this is a goodbye. goodbye again.")

if __name__ == "__main__":
    unittest.main()
